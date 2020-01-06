package broker

import (
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	osb "gitlab.bj.sensetime.com/service-providers/go-open-service-broker-client/v2"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var data = `
- name: clickhouse
  id: "558888d1-f00a-3176-01e4-05cd2d68eb18"
  description: "Provide clickhouse as a service through Open service Broker!"
  bindable: true
  planupdatable: true
  metadata:
    display_name: clickhouse service
    image_url: https://gitlab.bj.sensetime.com/uploads/-/system/project/avatar/14162/Clickhouse-logo-small.png
  plans:
  - name: default
    id: 2db0b31d-6912-4d24-8704-cfdf9b98af81
    description: The default plan for the clickhouse service
    free: true
    schemas:
      serviceinstance:
        create:
          parameters:
            image_version: v3.1
            nodes_per_rack: 2
            rack_num: 1
            data_capacity: 30Gi`

func TestNewClickHouseCluster(t *testing.T) {
	services := make([]osb.Service, 0)
	if err := yaml.Unmarshal([]byte(data), &services); err != nil {
		t.Fatalf(err.Error())
	}
	paras := services[0].Plans[0].Schemas.ServiceInstance.Create.Parameters
	planSpec := ParametersSpec{}
	err := mapstructure.Decode(paras, &planSpec)
	assert.Nil(t, err)

	userDefinedPara := map[string]interface{}{"name": "test", "namespace": "default"}

	instance := Instance{
		ID:        "mock",
		ServiceID: "mock",
		PlanID:    "mock",
		Params:    userDefinedPara,
	}

	meta := metav1.ObjectMeta{
		Name:      instance.Params["name"].(string),
		Namespace: instance.Params["namespace"].(string),
		Labels: map[string]string{
			InstanceID: instance.ID, //oss defined
			ServiceID:  instance.ServiceID,
			PlanID:     instance.PlanID,
		},
	}

	NewClickHouseCluster(&planSpec, meta)
	assert.Nil(t, err)
}

func TestBytesToStringSlice(t *testing.T) {
	in := `["abc", "def", "123"]`
	o := BytesToStringSlice([]byte(in))
	assert.Equal(t, o[0], "abc")
	assert.Equal(t, o[1], "def")
	assert.Equal(t, o[2], "123")
}
