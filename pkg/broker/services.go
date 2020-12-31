package broker

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"io/ioutil"
	"reflect"

	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/mackwong/clickhouse-operator/pkg/apis"
	"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/mitchellh/mapstructure"
	osb "gitlab.bj.sensetime.com/service-providers/go-open-service-broker-client/v2"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	clientrest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReadFromConfigMap(configPath string) (*[]osb.Service, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	services := make([]osb.Service, 0)
	err = yaml.Unmarshal(data, &services)
	if err != nil {
		return nil, err
	}

	for _, s := range services {
		for _, p := range s.Plans {
			if p.Schemas != nil && p.Schemas.ServiceInstance != nil {
				if p.Schemas.ServiceInstance.Create != nil {
					spec := ParametersSpec{}
					if err = mapstructure.Decode(p.Schemas.ServiceInstance.Create.Parameters, &spec); err != nil {
						return nil, err
					}
					p.Schemas.ServiceInstance.Create.Parameters = spec
				}
				if p.Schemas.ServiceInstance.Update != nil {
					spec := UpdateParametersSpec{}
					if err = mapstructure.Decode(p.Schemas.ServiceInstance.Update.Parameters, &spec); err != nil {
						return nil, err
					}
					p.Schemas.ServiceInstance.Update.Parameters = spec
				}
			}
		}
	}
	return &services, err
}

func GetClickHouseClient(kubeConfigPath string) (client.Client, *monclientv1.MonitoringV1Client, error) {
	var clientConfig *clientrest.Config
	var err error
	if kubeConfigPath == "" {
		clientConfig, err = clientrest.InClusterConfig()
		if err != nil {
			return nil, nil, err
		}
	} else {
		config, err := clientcmd.LoadFromFile(kubeConfigPath)
		if err != nil {
			return nil, nil, err
		}

		clientConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, nil, err
		}
	}

	s := scheme.Scheme
	//s := runtime.NewScheme()
	err = apis.AddToScheme(s)
	if err != nil {
		return nil, nil, err
	}
	err = v1beta1.AddToScheme(s)
	if err != nil {
		return nil, nil, err
	}

	cli, err := client.New(clientConfig, client.Options{
		Scheme: s,
	})
	if err != nil {
		return nil, nil, err
	}

	var ok bool
	var mClient *monclientv1.MonitoringV1Client
	dc := discovery.NewDiscoveryClientForConfigOrDie(clientConfig)
	if ok, err = k8sutil.ResourceExists(dc, "monitoring.coreos.com/v1", "ServiceMonitor"); err != nil {
		return nil, nil, err
	} else if ok {
		mClient = monclientv1.NewForConfigOrDie(clientConfig)
	}
	return cli, mClient, err
}

func NewClickHouseCluster(spec *ParametersSpec, meta metav1.ObjectMeta) *v1.ClickHouseCluster {
	clickhouse := &v1.ClickHouseCluster{
		ObjectMeta: meta,
		Spec:       spec.ToClickHouseClusterSpec(),
	}
	return clickhouse
}

type UpdateParametersSpec struct {
	//DeletePVC defines if the PVC must be deleted when the cluster is deleted
	//it is false by default
	DeletePVC bool `json:"deletePVC,omitempty"`

	//Shards count
	ShardsCount int32 `json:"shardsCount,omitempty"`

	//Replicas count
	ReplicasCount int32 `json:"replicasCount,omitempty"`

	Resources v1.ClickHouseResources `json:"resources,omitempty"`
}

type ParametersSpec v1.ClickHouseClusterSpec

func (p *ParametersSpec) merge(plan *ParametersSpec) {
	userElem := reflect.ValueOf(p).Elem()
	planElem := reflect.ValueOf(plan).Elem()
loop:
	for i := 0; i < planElem.NumField(); i++ {
		planField := planElem.Field(i)
		userField := userElem.Field(i)
		if userField.Kind() == reflect.Struct {
			for j := 0; j < userField.NumField(); j++ {
				if userField.Field(j).Kind() == reflect.Slice {
					continue loop
				}
			}
		}
		if userField.Interface() == reflect.Zero(userField.Type()).Interface() {
			userField.Set(reflect.ValueOf(planField.Interface()))
		}
	}
}

func (p *ParametersSpec) ToClickHouseClusterSpec() v1.ClickHouseClusterSpec {
	return v1.ClickHouseClusterSpec(*p)
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
