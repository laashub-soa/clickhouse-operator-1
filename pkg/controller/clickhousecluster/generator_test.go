package clickhousecluster

import (
	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type GeneratorTestSuite struct {
	suite.Suite
	g *Generator
}

func (g *GeneratorTestSuite) SetupTest() {
	chc := &v1.ClickHouseCluster{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fack",
			Namespace: "default",
		},
		Spec: v1.ClickHouseClusterSpec{
			ShardsCount:   2,
			ReplicasCount: 3,
		},
		Status: v1.ClickHouseClusterStatus{},
	}
	g.g = NewGenerator(chc)
}

func (g *GeneratorTestSuite) TestGenerateRemoteServersXML() {
	out := g.g.generateRemoteServersXML()
	g.T().Log(out)
}

func (g *GeneratorTestSuite) TestGenerateAllMacrosJson() {
	out := g.g.generateAllMacrosJson()
	if out == "" {
		g.T().Fatal("generate Macros error")
	}
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(GeneratorTestSuite))
}
