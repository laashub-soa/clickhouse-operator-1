package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1alpha1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/mackwong/clickhouse-operator/pkg/controller/clickhousecluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	osb "gitlab.bj.sensetime.com/service-providers/go-open-service-broker-client/v2"
	"gitlab.bj.sensetime.com/service-providers/osb-broker-lib/pkg/broker"
)

const (
	OperateTimeOut    = 30 * time.Second
	InstanceID        = "instance_name"
	InstanceName      = "instance_name"
	InstanceNamespace = "namespace"
	ServiceID         = "service_id"
	PlanID            = "plan_id"

	ProvisionOperation   = "Provision"
	DeprovisionOperation = "Deprovision"
	UpdateOperation      = "Update"
)

var (
	minSupportedVersion = "2.13"
	versionConstraint   = ">= " + minSupportedVersion
)

// NewCHCBrokerLogic represents the creation of a CHCBrokerLogic
func NewCHCBrokerLogic(KubeConfig string, o Options) (*CHCBrokerLogic, error) {
	glog.Infof("NewCHCBrokerLogic called.\n")

	services, err := ReadFromConfigMap(o.ServiceConfigPath)
	if err != nil {
		glog.Fatalf("can not load services config from %s, err: %s", o.ServiceConfigPath, err)
		return nil, err
	}
	if services == nil {
		err = fmt.Errorf("can not find any service from config")
		glog.Fatal(err.Error())
		return nil, err
	}

	cli, err := GetClickHouseClient(KubeConfig)
	if err != nil {
		glog.Fatalf("can not get a client, err: %s", err)
		return nil, err
	}
	return &CHCBrokerLogic{
		services:  services,
		cli:       cli,
		instances: make(map[string]*Instance, 10),
		bindings:  make(map[string]*BindingInfo, 10),
	}, nil
}

// CHCBrokerLogic is where stores the metadata and client of a service
type CHCBrokerLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	broker.Interface
	// Synchronize go routines.
	sync.RWMutex
	services  *[]osb.Service
	cli       client.Client
	instances map[string]*Instance
	bindings  map[string]*BindingInfo
}

var _ broker.Interface = &CHCBrokerLogic{}

func (b *CHCBrokerLogic) recoveryInstance(item *v1beta1.ServiceInstance) {
	instanceID := item.Spec.ExternalID
	serviceID := item.Spec.ClusterServiceClassName
	planID := item.Spec.ClusterServicePlanName
	parameters := make(map[string]interface{})

	if item.Spec.Parameters != nil {
		err := json.Unmarshal(item.Spec.Parameters.Raw, &parameters)
		if err != nil {
			glog.Errorf("unmarshal parameters err: %s", err)
			return
		}
	}

	if instanceID == "" || serviceID == "" || planID == "" {
		glog.V(5).Infoln("skip recovery serviceInstance, cuz all IDs are null")
		return
	}
	instance := Instance{
		ID:        instanceID,
		Name:      item.Name,
		Namespace: item.Namespace,
		ServiceID: serviceID,
		PlanID:    planID,
		Params:    parameters,
	}
	b.instances[instanceID] = &instance
	glog.V(5).Infof("recoveryInstance serviceInstance: %s\n", instanceID)
	return
}

func (b *CHCBrokerLogic) recoveryBinding(item *v1beta1.ServiceBinding) {
	secretName := item.Spec.SecretName
	namespace := item.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()
	secret := corev1.Secret{}
	if err := b.cli.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, &secret); err != nil {
		if errors.IsNotFound(err) {
			glog.Warningf("can not find secret")
			return
		}
		err = fmt.Errorf("can not get secret: %s, err: %s", secretName, err)
		glog.Errorf(err.Error())
		return
	}

	b.bindings[item.Spec.ExternalID] = &BindingInfo{
		User:     string(secret.Data["user"]),
		Password: string(secret.Data["password"]),
		Host:     BytesToStringSlice(secret.Data["host"]),
	}
	glog.V(5).Infof("recoveryInstance bindings: %s", item.Spec.ExternalID)
}

// Recovery instances data from restart
func (b *CHCBrokerLogic) Recovery() error {
	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()

	serviceInstanceList := v1beta1.ServiceInstanceList{}
	if err := b.cli.List(ctx, &serviceInstanceList, &client.ListOptions{}); err != nil {
		glog.Errorf("can not list service instance list, err: %s", err)
		return err
	}

	for _, item := range serviceInstanceList.Items {
		for _, service := range *b.services {
			if item.Spec.ClusterServiceClassName == service.ID {
				b.recoveryInstance(&item)
			}
		}
	}

	serviceBindingList := v1beta1.ServiceBindingList{}
	if err := b.cli.List(ctx, &serviceBindingList, &client.ListOptions{}); err != nil {
		glog.Errorf("can not list service instance list, err: %s", err)
		return err
	}

	for _, item := range serviceBindingList.Items {
		secretName := item.Spec.SecretName
		namespace := item.Namespace

		if secretName == "" {
			glog.Warningf("can not find secret in serviceBinding: %s/%s", namespace, item.Name)
			continue
		}

		for _, instance := range b.instances {
			if instance.Namespace == item.Namespace && instance.Name == item.Spec.InstanceRef.Name {
				b.recoveryBinding(&item)
			}
		}
	}
	return nil
}

// GetCatalog is to list all ServiceClasses and ServicePlans this broker supports
func (b *CHCBrokerLogic) GetCatalog(request *broker.RequestContext) (*broker.CatalogResponse, error) {
	glog.V(5).Infof("get request from GetCatalog: %s\n", toJson(request))
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: *b.services,
	}

	response.CatalogResponse = *osbResponse
	return response, nil
}

func (b *CHCBrokerLogic) doProvision(instance *Instance) (err error) {
	planSpec := ParametersSpec{}
	if err = mapstructure.Decode(instance.Params, &planSpec); err != nil {
		return
	}

loop:
	for _, service := range *b.services {
		if service.ID == instance.ServiceID {
			for _, plan := range service.Plans {
				if plan.ID == instance.PlanID {
					p := plan.Schemas.ServiceInstance.Create.Parameters.(ParametersSpec)
					planSpec.merge(&p)
					break loop
				}
			}
		}
	}

	meta := v1.ObjectMeta{
		Name:      instance.Name,
		Namespace: instance.Namespace,
		Labels: map[string]string{
			InstanceID: instance.ID, //oss defined
			ServiceID:  instance.ServiceID,
			PlanID:     instance.PlanID,
		},
	}

	chc := NewClickHouseCluster(&planSpec, meta)
	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()
	err = b.cli.Create(ctx, chc)
	if err != nil && errors.IsAlreadyExists(err) {
		err = b.cli.Update(ctx, chc)
	}
	if err != nil {
		glog.Errorf("create chc instance err: %s", err.Error())
	}
	return err
}

func (b *CHCBrokerLogic) doDeprovision(instance *Instance) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()
	clusterList := v1alpha1.ClickHouseClusterList{}

	labelSelect := labels.SelectorFromSet(map[string]string{InstanceID: instance.ID})
	err = b.cli.List(ctx, &clusterList, &client.ListOptions{
		LabelSelector: labelSelect,
	})

	for _, item := range clusterList.Items {
		err = b.cli.Delete(ctx, &item)
		if err != nil {
			glog.Errorf("delete clickhouse instance %s err: %s", instance.ID, err.Error())
			return err
		}
	}
	return nil
}

// Provision is to create a ServiceInstance, which actually also creates the CR of Clickhouse cluster, optionally, with given configuration
func (b *CHCBrokerLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	glog.V(5).Infof("get request from Provision: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()

	//https://github.com/openservicebrokerapi/servicebroker/blob/v2.15/spec.md#asynchronous-operations
	if !request.AcceptsIncomplete {
		return nil, asyncRequiredError
	}

	if request.ServiceID == "" || request.PlanID == "" ||
		request.Context[InstanceName] == nil || request.Context[InstanceNamespace] == nil ||
		!(b.validateServiceID(request.ServiceID) && b.validatePlanID(request.ServiceID, request.PlanID)) {
		errMsg := fmt.Sprintf("The request is malformed or missing mandatory data!")
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: &errMsg,
		}
	}

	response := broker.ProvisionResponse{
		ProvisionResponse: osb.ProvisionResponse{
			Async:        true,
			OperationKey: &[]osb.OperationKey{ProvisionOperation}[0],
		},
	}

	instance := &Instance{
		ID:        request.InstanceID,
		Name:      request.Context[InstanceName].(string),
		Namespace: request.Context[InstanceNamespace].(string),
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}

	// Check to see if this is the same instance
	if i := b.instances[request.InstanceID]; i != nil {
		if i.Match(instance) {
			response.Exists = true
			return &response, nil
		}
		description := "InstanceID in use"
		return nil, osb.HTTPStatusCodeError{
			StatusCode:  http.StatusConflict,
			Description: &description,
		}
	}

	err := b.doProvision(instance)
	if err != nil {
		description := err.Error()
		return nil, osb.HTTPStatusCodeError{
			StatusCode:  http.StatusServiceUnavailable,
			Description: &description,
		}
	}
	b.instances[request.InstanceID] = instance
	return &response, nil
}

// Deprovision is to delete the corresponding ServiceInstance, which also delete its cluster CR
func (b *CHCBrokerLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	glog.V(5).Infof("get request from Deprovision: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()

	if !request.AcceptsIncomplete {
		return nil, asyncRequiredError
	}

	response := broker.DeprovisionResponse{
		DeprovisionResponse: osb.DeprovisionResponse{
			Async:        true,
			OperationKey: &[]osb.OperationKey{DeprovisionOperation}[0],
		},
	}

	instance, ok := b.instances[request.InstanceID]
	if !ok {
		glog.Warningf("instance: %s is not exist", request.InstanceID)
		return &response, nil
	}

	err := b.doDeprovision(instance)
	if err != nil {
		description := err.Error()
		return nil, osb.HTTPStatusCodeError{
			StatusCode:  http.StatusServiceUnavailable,
			Description: &description,
		}
	}

	delete(b.instances, request.InstanceID)

	return &response, nil
}

func (b *CHCBrokerLogic) checkProvisionAction(instanceID string) (osb.LastOperationState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()
	clickHouseClusterList := v1alpha1.ClickHouseClusterList{}

	labelSelect := labels.SelectorFromSet(map[string]string{InstanceID: instanceID})
	err := b.cli.List(ctx, &clickHouseClusterList, &client.ListOptions{
		LabelSelector: labelSelect,
	})
	if err != nil {
		return osb.StateFailed, err
	}
	if len(clickHouseClusterList.Items) != 1 {
		return osb.StateFailed, fmt.Errorf("the num of find clickhousecluster is not 1")
	}
	if clickHouseClusterList.Items[0].Status.Phase == clickhousecluster.ClusterPhaseRunning {
		return osb.StateSucceeded, nil
	}
	return osb.StateInProgress, nil
}

func (b *CHCBrokerLogic) checkDeprovisionAction(instanceID string) (osb.LastOperationState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()
	clickHouseClusterList := v1alpha1.ClickHouseClusterList{}

	labelSelect := labels.SelectorFromSet(map[string]string{InstanceID: instanceID})
	err := b.cli.List(ctx, &clickHouseClusterList, &client.ListOptions{
		LabelSelector: labelSelect,
	})
	if err != nil {
		return osb.StateFailed, err
	}
	if len(clickHouseClusterList.Items) == 1 {
		return osb.StateInProgress, nil
	}
	if len(clickHouseClusterList.Items) == 0 {
		return osb.StateSucceeded, nil
	}
	return osb.StateFailed, fmt.Errorf("the num of find clickhousecluster is more than 1")
}

func (b *CHCBrokerLogic) checkUpdateAction(instanceID string) (osb.LastOperationState, error) {
	return b.checkProvisionAction(instanceID)
}

// LastOperation is to...
func (b *CHCBrokerLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	glog.V(5).Infof("get request from LastOperation: %s\n", toJson(request))
	if request.OperationKey == nil {
		return nil, osb.HTTPStatusCodeError{
			StatusCode:  http.StatusServiceUnavailable,
			Description: &[]string{"can not find operation key"}[0],
		}
	}
	var state osb.LastOperationState
	var err error
	switch *request.OperationKey {
	case ProvisionOperation:
		state, err = b.checkProvisionAction(request.InstanceID)
	case DeprovisionOperation:
		state, err = b.checkDeprovisionAction(request.InstanceID)
	case UpdateOperation:
		state, err = b.checkUpdateAction(request.InstanceID)
	default:
		err = osb.HTTPStatusCodeError{
			StatusCode:  http.StatusServiceUnavailable,
			Description: &[]string{"unknown operation key"}[0],
		}
	}
	return &broker.LastOperationResponse{
		LastOperationResponse: osb.LastOperationResponse{
			State: state,
		},
	}, err

}

var ccc int

func (b *CHCBrokerLogic) ExtensionLastOperation(request *osb.ExtensionLastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	ccc = ccc + 10
	desc := fmt.Sprintf("%d", ccc)
	operationKey := osb.OperationKey(*request.ActionID)
	if operationKey == "test" {
		if ccc > 50 {
			return &broker.LastOperationResponse{
				osb.LastOperationResponse{
					State:       osb.StateSucceeded,
					Description: &desc,
				},
			}, nil
		}
		return &broker.LastOperationResponse{
			osb.LastOperationResponse{
				State:       osb.StateInProgress,
				Description: &desc,
			},
		}, nil
	}
	return nil, nil
}

// Bind is to create a Binding, which also generates a user of ClickHouse, optionally, with given username and password
func (b *CHCBrokerLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	glog.V(5).Infof("get request from Bind: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()

	instance, ok := b.instances[request.InstanceID]
	if !ok {
		errMsg := fmt.Sprintf("can not find instance: %s", request.InstanceID)
		glog.Errorf(errMsg)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusNotFound,
			ErrorMessage: &errMsg,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()

	clusterList := v1alpha1.ClickHouseClusterList{}
	labelSelect := labels.SelectorFromSet(map[string]string{InstanceID: instance.ID})
	err := b.cli.List(ctx, &clusterList, &client.ListOptions{
		LabelSelector: labelSelect,
	})
	if err != nil {
		errMsg := fmt.Sprintf("can not list clickhouse err: %s\n", err)
		glog.Errorf(errMsg)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusServiceUnavailable,
			ErrorMessage: &errMsg,
		}
	}

	if len(clusterList.Items) != 1 {
		errMsg := fmt.Sprintf("find more than one clickhousecluster have same instance ID")
		glog.Errorf(errMsg)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusServiceUnavailable,
			ErrorMessage: &errMsg,
		}
	}

	host := getCHCServiceName(instance.Name, instance.Namespace)
	user, password := RandStringRunes(5), RandStringRunes(20)
	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: map[string]interface{}{
				"host":     []string{host},
				"user":     user,
				"password": password,
			},
			Async: false,
		},
	}
	b.bindings[request.BindingID] = &BindingInfo{
		User:     user,
		Password: password,
		Host:     []string{host},
	}

	return &response, nil
}

//Unbind is to delete a binding and the user it generated
func (b *CHCBrokerLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	glog.V(5).Infof("get request from Bind: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()
	return nil, nil
}

func (b *CHCBrokerLogic) Extension(request *osb.ExtensionRequest, c *broker.RequestContext) (*broker.ExtensionResponse, error) {
	operationKey := osb.OperationKey(request.ActionID)
	response := broker.ExtensionResponse{
		ExtensionResponse: osb.ExtensionResponse{
			Async:        true,
			OperationKey: &operationKey,
		},
		Exists: false,
	}
	return &response, nil
}

func (b *CHCBrokerLogic) UndoExtension(request *osb.UndoExtensionRequest, c *broker.RequestContext) (*broker.UndoExtensionResponse, error) {
	operationKey := osb.OperationKey(request.ActionID)
	response := broker.UndoExtensionResponse{
		UndoExtensionResponse: osb.UndoExtensionResponse{
			Async:        false,
			OperationKey: &operationKey,
		},
		Exists: false,
	}
	return &response, nil
}

func (b *CHCBrokerLogic) GetDocumentation(request *osb.GetDocumentationRequest, c *broker.RequestContext) (*broker.DocumentationResponse, error) {
	return &broker.DocumentationResponse{
		osb.GetDocumentationResponse{
			Documentation: "just for test",
		},
	}, nil
}

// Update is to update the CR of cluster
func (b *CHCBrokerLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	glog.V(5).Infof("get request from Update: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()

	if !request.AcceptsIncomplete {
		return nil, asyncRequiredError
	}

	instance, ok := b.instances[request.InstanceID]
	if !ok {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}
	instance.Params = request.Parameters

	namespace, ok := request.Context[InstanceNamespace].(string)
	if ok && namespace != instance.Namespace {
		errMsg := "can not change namespace"
		return nil, osb.HTTPStatusCodeError{
			ErrorMessage: &errMsg,
			StatusCode:   http.StatusBadRequest,
		}
	}

	err := b.doProvision(instance)
	if err != nil {
		description := err.Error()
		return nil, osb.HTTPStatusCodeError{
			StatusCode:  http.StatusServiceUnavailable,
			Description: &description,
		}
	}
	b.instances[request.InstanceID] = instance

	response := broker.UpdateInstanceResponse{
		UpdateInstanceResponse: osb.UpdateInstanceResponse{
			Async:        true,
			OperationKey: &[]osb.OperationKey{UpdateOperation}[0],
		},
	}
	response.Async = true
	return &response, nil
}

func (b *CHCBrokerLogic) ValidateBrokerAPIVersion(version string) error {
	if version == "" {
		errMsg, errDsp := "Precondition Failed",
			"Reject requests without X-Broker-API-Version header!"
		return osb.HTTPStatusCodeError{
			StatusCode:   http.StatusPreconditionFailed,
			ErrorMessage: &errMsg,
			Description:  &errDsp,
		}
	} else if !validateBrokerAPIVersion(version) {
		errMsg, errDsp := "Precondition Failed",
			fmt.Sprintf("API version %v is not supported by broker!", version)
		return osb.HTTPStatusCodeError{
			StatusCode:   http.StatusPreconditionFailed,
			ErrorMessage: &errMsg,
			Description:  &errDsp,
		}
	}
	return nil
}

//func (b *CHCBrokerLogic) validateParameters(parameters map[string]interface{}) error {
//	if _, ok := parameters["cluster_name"]; !ok {
//		return fmt.Errorf("have not assign cluster")
//	}
//	return nil
//}

func (b *CHCBrokerLogic) validateServiceID(serviceID string) bool {
	for _, s := range *b.services {
		if s.ID == serviceID {
			return true
		}
	}
	return false
}

func (b *CHCBrokerLogic) validatePlanID(serviceID, planID string) bool {
	for _, s := range *b.services {
		if s.ID == serviceID {
			for _, p := range s.Plans {
				if p.ID == planID {
					return true
				}
			}
		}
	}
	return false
}
