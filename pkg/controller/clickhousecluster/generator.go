package clickhousecluster

import (
	"encoding/json"
	"fmt"
	clickhousev1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/sirupsen/logrus"
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
	InitContainerName       = "clickhouse-init"

	filenameRemoteServersXML = "remote_servers.xml"
	filenameAllMacrosJSON    = "all-macros.json"
	//filenameUsersXML         = "users.xml"
	filenameZookeeperXML = "zookeeper.xml"
	filenameSettingsXML  = "settings.xml"

	dirPathConfigd = "/etc/clickhouse-server/config.d/"
	//dirPathUsersd  = "/etc/clickhouse-server/users.d/"
	dirPathConfd = "/etc/clickhouse-server/conf.d/"

	macrosTemplate = `
<yandex>
        <macros>
            <cluster>%s</cluster>
            <shard>%d</shard>
            <replica>%s</replica>
        </macros>
</yandex>`
)

type Generator struct {
	rcc *ReconcileClickHouseCluster
	cc  *clickhousev1.ClickHouseCluster
}

func NewGenerator(rcc *ReconcileClickHouseCluster, cc *clickhousev1.ClickHouseCluster) *Generator {
	return &Generator{rcc: rcc, cc: cc}
}

func (g *Generator) labelsForStatefulSet(shardID int) map[string]string {
	return map[string]string{
		"clickhouse-cluster": g.cc.Name,
		"shard-id":           fmt.Sprintf("%d", shardID),
	}
}

func (g *Generator) labelsForCluster() map[string]string {
	return map[string]string{
		"clickhouse-cluster": g.cc.Name,
	}
}

func (g *Generator) ownerReference() []metav1.OwnerReference {
	ref := metav1.OwnerReference{
		APIVersion: g.cc.APIVersion,
		Kind:       g.cc.Kind,
		Name:       g.cc.Name,
		UID:        g.cc.UID,
	}
	return []metav1.OwnerReference{ref}
}

func (g *Generator) marosEmptyDirName() string {
	return fmt.Sprintf("clickhouse-%s-maros", g.cc.Name)
}

func (g *Generator) commonConfigMapName() string {
	return fmt.Sprintf("clickhouse-%s-common-config", g.cc.Name)
}

func (g *Generator) userConfigMapName() string {
	return fmt.Sprintf("clickhouse-%s-user-config", g.cc.Name)
}

func (g *Generator) macrosConfigMapName() string {
	return fmt.Sprintf("clickhouse-%s-macros-config", g.cc.Name)
}

func (g *Generator) statefulSetName(shardID int) string {
	return fmt.Sprintf("clickhouse-%s-%d", g.cc.Name, shardID)
}

func (g *Generator) serviceName(shardID int) string {
	return g.statefulSetName(shardID)
}

func (g *Generator) FQDN(shardID, replicasID int, namespace string) string {
	statefulset := g.statefulSetName(shardID)
	serviceName := g.serviceName(shardID)
	return fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", statefulset, replicasID, serviceName, namespace)
}

func (g *Generator) generateRemoteServersXML() string {
	shards := make([]Shard, g.cc.Spec.ShardsCount)
	for i := range shards {
		replicas := make([]Replica, g.cc.Spec.ReplicasCount)
		for j := range replicas {
			replicas[j].Host = g.FQDN(i, j, g.cc.Namespace)
			replicas[j].Port = chDefaultClientPortNumber
		}
		shards[i].InternalReplication = true
		shards[i].Replica = replicas
	}

	servers := RemoteServers{RemoteServer: map[string]Cluster{
		g.cc.Name: {shards},
	}}
	return ParseXML(servers)
}

func (g *Generator) generateZookeeperXML() string {
	zk := Zookeeper{Zookeeper: g.cc.Spec.Zookeeper}
	return ParseXML(zk)
}

func (g *Generator) generateSettingsXML() string {
	return g.cc.Spec.CustomSettings
}

func (g *Generator) generateAllMacrosJson() string {
	macros := make(map[string]string)
	var shardsCount = int(g.cc.Spec.ShardsCount)
	var replicasCount = int(g.cc.Spec.ReplicasCount)
	for i := 0; i < shardsCount; i++ {
		for j := 0; j < replicasCount; j++ {
			replica := fmt.Sprintf("%s-%d", g.statefulSetName(i), j)
			macros[replica] = fmt.Sprintf(macrosTemplate, g.cc.Name, i, replica)
		}
	}
	out, err := json.MarshalIndent(macros, " ", "")
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Marshal error")
		return ""
	}
	return string(out)
}

func (g *Generator) GenerateCommonConfigMap() *corev1.ConfigMap {

	data := map[string]string{
		filenameRemoteServersXML: g.generateRemoteServersXML(),
		filenameAllMacrosJSON:    g.generateAllMacrosJson(),
		filenameSettingsXML:      g.generateSettingsXML(),
		filenameZookeeperXML:     g.generateZookeeperXML(),
	}
	for filename, content := range g.rcc.defaultConfig.GetDefaultXMLConfig() {
		data[filename] = content
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            g.commonConfigMapName(),
			Namespace:       g.cc.Namespace,
			Labels:          g.labelsForCluster(),
			OwnerReferences: g.ownerReference(),
		},
		// Data contains several sections which are to be several xml chopConfig files
		Data: data,
	}
}

func (g *Generator) generateUserConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            g.userConfigMapName(),
			Namespace:       g.cc.Namespace,
			Labels:          g.labelsForCluster(),
			OwnerReferences: g.ownerReference(),
		},
		// Data contains several sections which are to be several xml chopConfig files
		Data: map[string]string{},
	}
}

func (g *Generator) generateService(shardID int) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            g.serviceName(shardID),
			Namespace:       g.cc.Namespace,
			Labels:          g.labelsForStatefulSet(shardID),
			OwnerReferences: g.ownerReference(),
		},
		Spec: corev1.ServiceSpec{
			// ClusterIP: templateDefaultsServiceClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     chDefaultHTTPPortName,
					Port:     chDefaultHTTPPortNumber,
					Protocol: "TCP",
				},
				{
					Name:     chDefaultClientPortName,
					Port:     chDefaultClientPortNumber,
					Protocol: "TCP",
				},
			},
			Selector:  g.labelsForStatefulSet(shardID),
			ClusterIP: "None",
			Type:      "ClusterIP",
		},
	}
}

func (g *Generator) setupStatefulSetPodTemplate(statefulset *appsv1.StatefulSet, shardID int) {
	statefulset.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   statefulset.Name,
			Labels: g.labelsForStatefulSet(shardID),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{},
			Volumes:    []corev1.Volume{},
		},
	}
	statefulset.Spec.Template.Spec.InitContainers = []corev1.Container{
		{
			Name:  InitContainerName,
			Image: g.cc.Spec.InitImage,
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath:  "metadata.name",
							APIVersion: "clickhousev1",
						},
					},
				},
			},
		},
	}
	statefulset.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  ClickHouseContainerName,
			Image: g.cc.Spec.Image,
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
			},
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
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{
						"NET_ADMIN",
						"SYS_NICE",
					},
				},
			},
		},
	}

	// Add all ConfigMap objects as Volume objects of type ConfigMap
	statefulset.Spec.Template.Spec.Volumes = append(
		statefulset.Spec.Template.Spec.Volumes,
		newVolumeForConfigMap(g.commonConfigMapName()),
		newVolumeForEmptyDir(g.marosEmptyDirName()),
		//newVolumeForConfigMap(g.userConfigMapName()),
		//newVolumeForConfigMap(g.macrosConfigMapName()),
	)

	// And reference these Volumes in each Container via VolumeMount
	// So Pod will have ConfigMaps mounted as Volumes
	for i := range statefulset.Spec.Template.Spec.Containers {
		// Convenience wrapper
		container := &statefulset.Spec.Template.Spec.Containers[i]
		// Append to each Container current VolumeMount's to VolumeMount's declared in template
		container.VolumeMounts = append(
			container.VolumeMounts,
			newVolumeMount(g.commonConfigMapName(), dirPathConfigd),
			newVolumeMount(g.marosEmptyDirName(), dirPathConfd),
			//newVolumeMount(g.userConfigMapName(), dirPathUsersd),
			//newVolumeMount(g.macrosConfigMapName(), dirPathConfd),
		)
	}

	for i := range statefulset.Spec.Template.Spec.InitContainers {
		// Convenience wrapper
		container := &statefulset.Spec.Template.Spec.InitContainers[i]
		// Append to each Container current VolumeMount's to VolumeMount's declared in template
		container.VolumeMounts = append(
			container.VolumeMounts,
			newVolumeMount(g.commonConfigMapName(), dirPathConfigd),
			newVolumeMount(g.marosEmptyDirName(), dirPathConfd),
			//newVolumeMount(g.userConfigMapName(), dirPathUsersd),
			//newVolumeMount(g.macrosConfigMapName(), dirPathConfd),
		)
	}
}

//TODO
func (g *Generator) setupStatefulSetVolumeClaimTemplates(statefulset *appsv1.StatefulSet) {

}

func (g *Generator) generateStatefulSet(shardID int) *appsv1.StatefulSet {
	// Create apps.StatefulSet object
	replicasNum := g.cc.Spec.ReplicasCount
	// StatefulSet has additional label - ZK config fingerprint
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            g.statefulSetName(shardID),
			Namespace:       g.cc.Namespace,
			Labels:          g.labelsForStatefulSet(shardID),
			OwnerReferences: g.ownerReference(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicasNum,
			//ServiceName: g.serviceName(shardID),
			Selector: &metav1.LabelSelector{
				MatchLabels: g.labelsForStatefulSet(shardID),
			},
			ServiceName: g.serviceName(shardID),
			// IMPORTANT
			// VolumeClaimTemplates are to be setup later
			VolumeClaimTemplates: nil,

			// IMPORTANT
			// Template is to be setup later
			Template: corev1.PodTemplateSpec{},
		},
	}

	g.setupStatefulSetPodTemplate(statefulSet, shardID)
	g.setupStatefulSetVolumeClaimTemplates(statefulSet)

	return statefulSet
}

// newVolumeForConfigMap returns corev1.Volume object with defined name
func newVolumeForConfigMap(name string) corev1.Volume {
	var defaultMode int32 = 420
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				DefaultMode: &defaultMode,
			},
		},
	}
}

func newVolumeForEmptyDir(name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

// newVolumeMount returns corev1.VolumeMount object with name and mount path
func newVolumeMount(name, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	}
}
