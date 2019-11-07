package clickhousecluster

import (
	"fmt"
	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// ClickHouse open ports
	chDefaultHTTPPortName          = "http"
	chDefaultHTTPPortNumber        = 8123
	chDefaultClientPortName        = "client"
	chDefaultClientPortNumber      = 9000
	chDefaultInterServerPortName   = "interserver"
	chDefaultInterServerPortNumber = 9009

	ClickHouseContainerName = "clickhouse"

	filenameRemoteServersXML = "remote_servers.xml"
	filenameUsersXML         = "users.xml"
	filenameZookeeperXML     = "zookeeper.xml"
	filenameSettingsXML      = "settings.xml"
)

type Generator struct {
	clickHouseCluster *v1.ClickHouseCluster
}

func NewGenerator(chc *v1.ClickHouseCluster) *Generator {
	return &Generator{clickHouseCluster: chc}
}

func (g *Generator) labels() map[string]string {
	return map[string]string{
		"clickhousecluster": g.clickHouseCluster.Name,
	}
}

func (g *Generator) commonConfigMapName() string {
	return fmt.Sprintf("clickhouse-%s-common-config", g.clickHouseCluster.Name)
}

func (g *Generator) statefulsetName() string {
	return fmt.Sprintf("clickhouse-%s", g.clickHouseCluster.Name)
}

func (g *Generator) serviceName() string {
	return g.statefulsetName()
}

func (g *Generator) generateRemoteServersXML() string {
	shards := make([]Shard, g.clickHouseCluster.Spec.ShardsCount)
	statefulset := g.statefulsetName()
	index := 0
	for i := range shards {
		replicas := make([]Replica, g.clickHouseCluster.Spec.ReplicasCount)
		for j := range replicas {
			replicas[j].Host = fmt.Sprintf("%s-%d", statefulset, index)
			replicas[j].Port = chDefaultClientPortNumber
			index++
		}
		shards[i].InternalReplication = true
		shards[i].Replica = replicas
	}

	servers := YandexRemoteServers{
		Yandex: RemoteServers{RemoteServer: map[string]Cluster{
			g.clickHouseCluster.Name: {shards},
		}},
	}
	return ParseXML(servers)
}

func (g *Generator) generateZookeeperXML() string {
	return ""
}

func (g *Generator) generateSettingsXML() string {
	return ""
}

func (g *Generator) GenerateCommonConfigMap() *corev1.ConfigMap {

	data := map[string]string{
		filenameRemoteServersXML: g.generateRemoteServersXML(),
		filenameSettingsXML:      g.generateSettingsXML(),
		filenameZookeeperXML:     g.generateZookeeperXML(),
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.commonConfigMapName(),
			Namespace: g.clickHouseCluster.Namespace,
			Labels:    g.labels(),
		},
		// Data contains several sections which are to be several xml chopConfig files
		Data: data,
	}
}

func (g *Generator) CreateConfigMapUsers() *corev1.ConfigMap {
	return nil
}

func (g *Generator) GenerateService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.clickHouseCluster.Name,
			Namespace: g.clickHouseCluster.Namespace,
			Labels:    g.labels(),
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
			Selector: g.labels(),
			Type:     "ClusterIP",
		},
	}
}

func (g *Generator) setupStatefulSetPodTemplate(statefulset *appsv1.StatefulSet) {
	container := corev1.Container{
		Name:  ClickHouseContainerName,
		Image: g.clickHouseCluster.Spec.Image,
		Ports: []corev1.ContainerPort{
			{
				Name:          chDefaultHTTPPortName,
				ContainerPort: chDefaultHTTPPortNumber,
			},
			{
				Name:          chDefaultClientPortName,
				ContainerPort: chDefaultClientPortNumber,
			},
			{
				Name:          chDefaultInterServerPortName,
				ContainerPort: chDefaultInterServerPortNumber,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ping",
					Port: intstr.Parse(chDefaultHTTPPortName),
				},
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
		},
	}
	statefulset.Spec.Template.Spec.Containers = []corev1.Container{container}
}

func (g *Generator) GenerateStatefulset() *appsv1.StatefulSet {
	// Create apps.StatefulSet object
	replicasNum := g.clickHouseCluster.Spec.ShardsCount * g.clickHouseCluster.Spec.ReplicasCount
	// StatefulSet has additional label - ZK config fingerprint
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.statefulsetName(),
			Namespace: g.clickHouseCluster.Namespace,
			Labels:    g.labels(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicasNum,
			ServiceName: g.serviceName(),
			Selector: &metav1.LabelSelector{
				MatchLabels: g.labels(),
			},
			// IMPORTANT
			// VolumeClaimTemplates are to be setup later
			VolumeClaimTemplates: nil,

			// IMPORTANT
			// Template is to be setup later
			Template: corev1.PodTemplateSpec{},
		},
	}

	g.setupStatefulSetPodTemplate(statefulSet)
	//g.setupStatefulSetVolumeClaimTemplates(statefulSet, host)

	return statefulSet
}
