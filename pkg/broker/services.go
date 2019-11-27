package broker

import (
	"io/ioutil"
	"reflect"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/mackwong/clickhouse-operator/pkg/apis"
	"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/mitchellh/mapstructure"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			if p.Schemas != nil && p.Schemas.ServiceInstance != nil && p.Schemas.ServiceInstance.Create != nil {
				spec := ParametersSpec{}
				err = mapstructure.Decode(p.Schemas.ServiceInstance.Create.Parameters, &spec)
				if err != nil {
					return nil, err
				}
				p.Schemas.ServiceInstance.Create.Parameters = spec
			}
		}
	}
	return &services, err
}

func GetClickHouseClient(kubeConfigPath string) (client.Client, error) {
	var clientConfig *clientrest.Config
	var err error
	if kubeConfigPath == "" {
		clientConfig, err = clientrest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config, err := clientcmd.LoadFromFile(kubeConfigPath)
		if err != nil {
			return nil, err
		}

		clientConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	s := scheme.Scheme
	//s := runtime.NewScheme()
	err = apis.AddToScheme(s)
	if err != nil {
		return nil, err
	}
	err = v1beta1.AddToScheme(s)
	if err != nil {
		return nil, err
	}

	cli, err := client.New(clientConfig, client.Options{
		Scheme: s,
	})
	return cli, err
}

func NewClickHouseCluster(spec *ParametersSpec, meta metav1.ObjectMeta) (*v1.ClickHouseCluster, error) {
	clickhouse := &v1.ClickHouseCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClickHouseCluster",
			APIVersion: "service.diamond.sensetime.com/v1",
		},
		ObjectMeta: meta,
		Spec:       spec.ToClickHouseClusterSpec(),
	}
	return clickhouse, nil
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
