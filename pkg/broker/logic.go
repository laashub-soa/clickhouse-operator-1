package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1alpha1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"reflect"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

const (
	OperateTimeOut    = 30 * time.Second
	InstanceID        = "instance_name"
	InstanceName      = "instance_name"
	InstanceNamespace = "namespace"
	ServiceID         = "service_id"
	PlanID            = "plan_id"
)

var (
	minSupportedVersion = "2.13"
	versionConstraint   = ">= " + minSupportedVersion
)

// NewBusinessLogic represents the creation of a BusinessLogic
func NewBusinessLogic(KubeConfig string, o Options) (*BusinessLogic, error) {
	glog.Infof("NewBusinessLogic called.\n")

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
	return &BusinessLogic{
		async:     o.Async,
		services:  services,
		cli:       cli,
		instances: make(map[string]*Instance, 10),
		bindings:  make(map[string]*BindingInfo, 10),
	}, nil
}

// BusinessLogic is where stores the metadata and client of a service
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	broker.Interface
	async bool
	// Synchronize go routines.
	sync.RWMutex
	services  *[]osb.Service
	cli       client.Client
	instances map[string]*Instance
	bindings  map[string]*BindingInfo
}

var _ broker.Interface = &BusinessLogic{}

// Recovery instances data from restart
func (b *BusinessLogic) Recovery() error {
	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()

	serviceInstanceList := v1beta1.ServiceInstanceList{}
	if err := b.cli.List(ctx, &serviceInstanceList, &client.ListOptions{}); err != nil {
		glog.Errorf("can not list service instance list, err: %s", err)
		return err
	}

	for _, item := range serviceInstanceList.Items {
		var find bool
		for _, service := range *b.services {
			if item.Spec.ClusterServiceClassName == service.ID {
				find = true
				break
			}
		}
		if !find {
			continue
		}
		instanceID := item.Spec.ExternalID
		serviceID := item.Spec.ClusterServiceClassName
		planID := item.Spec.ClusterServicePlanName
		parameters := make(map[string]interface{})

		if item.Spec.Parameters != nil {
			err := json.Unmarshal(item.Spec.Parameters.Raw, &parameters)
			if err != nil {
				glog.Errorf("unmarshal parameters err: %s", err)
				return err
			}
		}

		if instanceID == "" || serviceID == "" || planID == "" {
			glog.V(5).Infoln("skip recovery serviceInstance, cuz all IDs are null")
			continue
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
		glog.V(5).Infof("recovery serviceInstance: %s\n", instanceID)
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

		var find bool
		for _, instance := range b.instances {
			if instance.Namespace == item.Namespace && instance.Name == item.Spec.InstanceRef.Name {
				find = true
				break
			}
		}
		if !find {
			continue
		}

		secret := corev1.Secret{}
		if err := b.cli.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, &secret); err != nil {
			if errors.IsNotFound(err) {
				glog.Warningf("can not find secret")
				continue
			}
			err = fmt.Errorf("can not get secret: %s, err: %s", secretName, err)
			glog.Errorf(err.Error())
			return err
		}

		b.bindings[item.Spec.ExternalID] = &BindingInfo{
			User:     string(secret.Data["user"]),
			Password: string(secret.Data["password"]),
			Host:     BytesToStringSlice(secret.Data["host"]),
		}
		glog.V(5).Infof("recovery bindings: %s", item.Spec.ExternalID)
	}
	return nil
}

// GetCatalog is to list all ServiceClasses and ServicePlans this broker supports
func (b *BusinessLogic) GetCatalog(request *broker.RequestContext) (*broker.CatalogResponse, error) {
	glog.V(5).Infof("get request from GetCatalog: %s\n", toJson(request))
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: *b.services,
	}

	response.CatalogResponse = *osbResponse
	return response, nil
}

func (b *BusinessLogic) doProvision(instance *Instance) (err error) {
	planSpec := ParametersSpec{}
	err = mapstructure.Decode(instance.Params, &planSpec)
	if err != nil {
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

	clickhouse, err := NewClickHouseCluster(&planSpec, meta)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), OperateTimeOut)
	defer cancel()
	err = b.cli.Create(ctx, clickhouse)
	if err != nil && errors.IsAlreadyExists(err) {
		err = b.cli.Update(ctx, clickhouse)
	}
	if err != nil {
		glog.Errorf("create clickhouse instance err: %s", err.Error())
	}
	return err
}

func (b *BusinessLogic) doDeprovision(instance *Instance) (err error) {
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
func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	glog.V(5).Infof("get request from Provision: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()

	if request.ServiceID == "" || request.PlanID == "" ||
		request.Context[InstanceName] == nil || request.Context[InstanceNamespace] == nil ||
		!(b.validateServiceID(request.ServiceID) && b.validatePlanID(request.ServiceID, request.PlanID)) {
		errMsg := fmt.Sprintf("The request is malformed or missing mandatory data!")
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: &errMsg,
		}
	}

	response := broker.ProvisionResponse{}

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

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

// Deprovision is to delete the corresponding ServiceInstance, which also delete its cluster CR
func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	glog.V(5).Infof("get request from Deprovision: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
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

// LastOperation is to...
func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	glog.V(5).Infof("get request from LastOperation: %s\n", toJson(request))
	return &broker.LastOperationResponse{}, nil
}

// Bind is to create a Binding, which also generates a user of ClickHouse, optionally, with given username and password
func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
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
	return nil, nil
}

//Unbind is to delete a binding and the user it generated
func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	glog.V(5).Infof("get request from Bind: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()
	return nil, nil
}

// Update is to update the CR of cluster
func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	glog.V(5).Infof("get request from Update: %s\n", toJson(request))
	b.Lock()
	defer b.Unlock()

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

	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}
	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
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

//func (b *BusinessLogic) validateParameters(parameters map[string]interface{}) error {
//	if _, ok := parameters["cluster_name"]; !ok {
//		return fmt.Errorf("have not assign cluster")
//	}
//	return nil
//}

func (b *BusinessLogic) validateServiceID(serviceID string) bool {
	for _, s := range *b.services {
		if s.ID == serviceID {
			return true
		}
	}
	return false
}

func (b *BusinessLogic) validatePlanID(serviceID, planID string) bool {
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

type BindingInfo struct {
	User     string
	Password string
	Host     []string
}

type Instance struct {
	ID        string
	Name      string
	Namespace string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
}

func (i *Instance) Match(other *Instance) bool {
	return reflect.DeepEqual(i, other)
}

func validateBrokerAPIVersion(version string) bool {
	c, _ := semver.NewConstraint(versionConstraint)
	v, _ := semver.NewVersion(version)
	// Check if the version meets the constraints. The a variable will be true.
	return c.Check(v)
}
