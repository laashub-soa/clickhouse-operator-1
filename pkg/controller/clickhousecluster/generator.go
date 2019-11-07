package clickhousecluster

import (
	"fmt"
	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ClickHouse open ports
	chDefaultHTTPPortName          = "http"
	chDefaultHTTPPortNumber        = 8123
	chDefaultClientPortName        = "client"
	chDefaultClientPortNumber      = 9000
	chDefaultInterServerPortName   = "interserver"
	chDefaultInterServerPortNumber = 9009

	filenameRemoteServersXML = "remote_servers.xml"
	filenameUsersXML         = "users.xml"
	filenameZookeeperXML     = "zookeeper.xml"
	filenameSettingsXML      = "settings.xml"
)

type Generator struct {
	clickHouseCluster *v1.ClickHouseCluster
}

func NewCreator() *Generator {
	return &Generator{}
}

func (c *Generator) labels() map[string]string {
	return map[string]string{
		"clickhousecluster": c.clickHouseCluster.Name,
	}
}

func (c *Generator) commonConfigMapName() string {
	return fmt.Sprintf("clickhouse-%s-common-config", c.clickHouseCluster.Name)
}

func (c *Generator) statefulsetName() string {
	return fmt.Sprintf("clickhouse-%s", c.clickHouseCluster.Name)
}

func (c *Generator) generateRemoteServersXML() string {
	return ""
	//servers := YandexRemoteServers{
	//	Yandex:RemoteServers{RemoteServer: map[string]Cluster{
	//		c.clickHouseCluster.Name: Shard{
	//			InternalReplication: false,
	//			Replica:             Replica{
	//				Host:
	//			},
	//		},
	//	}},
	//}
}

func (c *Generator) generateZookeeperXML() string {
	return ""
}

func (c *Generator) generateSettingsXML() string {
	return ""
}

func (c *Generator) GenerateCommonConfigMap() *corev1.ConfigMap {

	data := map[string]string{
		filenameRemoteServersXML: c.generateRemoteServersXML(),
		filenameSettingsXML:      c.generateSettingsXML(),
		filenameZookeeperXML:     c.generateZookeeperXML(),
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.commonConfigMapName(),
			Namespace: c.clickHouseCluster.Namespace,
			Labels:    c.labels(),
		},
		// Data contains several sections which are to be several xml chopConfig files
		Data: data,
	}
}

func (c *Generator) CreateConfigMapUsers() *corev1.ConfigMap {
	return nil
}

func (c *Generator) GenerateService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.clickHouseCluster.Name,
			Namespace: c.clickHouseCluster.Namespace,
			Labels:    c.labels(),
		},
		Spec: corev1.ServiceSpec{
			// ClusterIP: templateDefaultsServiceClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: chDefaultHTTPPortName,
					Port: chDefaultHTTPPortNumber,
				},
				{
					Name: chDefaultClientPortName,
					Port: chDefaultClientPortNumber,
				},
			},
			Selector: c.labels(),
			Type:     "ClusterIP",
		},
	}
}

func (c *Generator) GenerateStatefulset() *appsv1.StatefulSet {
	return nil
}
