package broker

import (
	"context"
	"fmt"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os/user"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"testing"
)

var (
	instanceID = "12345678-9012-2345-6789-ac9694d96d5b"
	planID     = "2db0b31d-6912-4d24-8704-cfdf9b98af81"
	serviceID  = "137a3ded-59ab-4ece-bbda-9cfff850a1f3"
)

type LogicTestSuite struct {
	suite.Suite
	logic     *BusinessLogic
	namespace string
}

//TODO: 利用runner的k8s进行测试
func (s *LogicTestSuite) SetupSuite() {
	var err error
	u, _ := user.Current()
	kubeconfig := path.Join(u.HomeDir, ".kube", "config")
	o := Options{
		ServiceConfigPath: "../../clickhouse.yaml",
		Async:             false,
	}
	s.logic, err = NewBusinessLogic(kubeconfig, o)
	assert.Nil(s.T(), err)

	s.namespace = fmt.Sprintf("test-%s", RandStringRunes(5))
	err = s.logic.cli.Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.namespace,
		},
	})
	s.T().Logf("create namespace: %s", s.namespace)
	if err != nil {
		assert.FailNow(s.T(), err.Error())
	}
}

func (s *LogicTestSuite) TearDownSuite() {
	var _ = s.logic.cli.Delete(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.namespace,
		},
	}, client.GracePeriodSeconds(0))
	s.T().Logf("delete namespace: %s", s.namespace)
}

func (s *LogicTestSuite) TestA_Recovery() {
	err := s.logic.Recovery()
	assert.Nil(s.T(), err)
}

func (s *LogicTestSuite) TestB_GetCatalog() {
	_, err := s.logic.GetCatalog(nil)
	assert.Nil(s.T(), err)
}

func (s *LogicTestSuite) TestC_Provision() {
	request := &osb.ProvisionRequest{
		InstanceID: instanceID,
		PlanID:     planID,
		ServiceID:  serviceID,
		Parameters: map[string]interface{}{
			"cluster_name": "test",
		},
		Context: map[string]interface{}{
			"namespace": s.namespace,
		},
	}

	_, err := s.logic.Provision(request, &broker.RequestContext{
		Request: &http.Request{
			Header: map[string][]string{
				"X-Broker-Api-Version": {"2.14"},
			},
		},
	})
	assert.Condition(s.T(), func() (success bool) {
		return err == nil || strings.Contains(err.Error(), "already exists")
	})
}

func (s *LogicTestSuite) TestD_Bind() {
	request := &osb.ProvisionRequest{
		InstanceID: instanceID,
		PlanID:     planID,
		ServiceID:  serviceID,
		Parameters: map[string]interface{}{
			"cluster_name": "test",
		},
		Context: map[string]interface{}{
			"namespace": s.namespace,
		},
	}
	_, err := s.logic.Provision(request, &broker.RequestContext{
		Request: &http.Request{
			Header: map[string][]string{
				"X-Broker-Api-Version": {"2.14"},
			},
		},
	})
	assert.Condition(s.T(), func() (success bool) {
		return err == nil || strings.Contains(err.Error(), "already exists")
	})
}

func (s *LogicTestSuite) TestE_UnBind() {
	request := &osb.UnbindRequest{
		InstanceID: instanceID,
		PlanID:     planID,
		ServiceID:  serviceID,
		BindingID:  "255370ad-c24a-11e9-88a2-6a2138d4a3dd",
	}
	_, err := s.logic.Unbind(request, &broker.RequestContext{
		Request: &http.Request{
			Header: map[string][]string{
				"X-Broker-Api-Version": {"2.14"},
			},
		}})
	assert.Condition(s.T(), func() (success bool) {
		return err == nil || (err != nil && strings.Contains(err.Error(), "failed to resolve any of the provided hostnames"))
	})
}

func (s *LogicTestSuite) TestF_Deprovision() {
	request := &osb.DeprovisionRequest{
		InstanceID: instanceID,
		PlanID:     planID,
		ServiceID:  serviceID,
	}
	_, err := s.logic.Deprovision(request, &broker.RequestContext{
		Request: &http.Request{
			Header: map[string][]string{
				"X-Broker-Api-Version": {"2.14"},
			},
		}})
	assert.Nil(s.T(), err)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(LogicTestSuite))
}
