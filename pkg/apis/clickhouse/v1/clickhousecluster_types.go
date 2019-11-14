package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationLastApplied string = "clickhouse.sensetime.com/last-applied-configuration"
)

// ClickHouseClusterSpec defines the desired state of ClickHouseCluster
// +k8s:openapi-gen=true
type ClickHouseClusterSpec struct {
	//ClickHouse Docker image
	Image string `json:"image,omitempty"`

	//ClickHouse init  image
	InitImage string `json:"initImage,omitempty"`

	//DeletePVC defines if the PVC must be deleted when the cluster is deleted
	//it is false by default
	DeletePVC bool `json:"deletePVC,omitempty"`

	//Shards count
	ShardsCount int32 `json:"shardsCount,omitempty"`

	//Replicas count
	ReplicasCount int32 `json:"replicasCount,omitempty"`

	//Zookeeper config
	Zookeeper *ZookeeperConfig `json:"zookeeper,omitempty"`

	//Custom defined XML settings, like <yandex>somethine</yandex>
	CustomSettings string

	//The storage capacity
	DataCapacity string `json:"dataCapacity,omitempty"`

	//Define StorageClass for Persistent Volume Claims in the local storage.
	DataStorageClass string `json:"dataStorageClass,omitempty"`

	//User defined pod spec
	//PodSpec *corev1.PodSpec `json:"podSpec,omitempty"`
}

// Remove subresources, cuz https://github.com/kubernetes/kubectl/issues/564
// ClickHouseClusterStatus defines the observed state of ClickHouseCluster
// +k8s:openapi-gen=true
type ClickHouseClusterStatus struct {
	Phase string `json:"phase,omitempty"`
	//CassandraRackStatusList list les Status pour chaque Racks
	ShardStatus map[string]*ShardStatus `json:"cassandraRackStatus,omitempty"`
}

//ShardStatus defines states of Clickhouse for 1 shard (1 statefulset)
type ShardStatus struct {
	// Phase indicates the state this Clickhouse cluster jumps in.
	// Phase goes as one way as below:
	//   Initial -> Running <-> updating
	Phase string `json:"phase,omitempty"`
}

// ZookeeperConfig defines zookeeper
// Refers to
// https://clickhouse.yandex/docs/en/single/index.html?#server-settings_zookeeper
type ZookeeperConfig struct {
	Nodes              []ZookeeperNode `json:"nodes,omitempty"                yaml:"nodes"  xml:"nodes"`
	SessionTimeoutMs   int             `json:"session_timeout_ms,omitempty"   yaml:"session_timeout_ms" xml:"session_timeout_ms"`
	OperationTimeoutMs int             `json:"operation_timeout_ms,omitempty" yaml:"operation_timeout_ms" xml:"operation_timeout_ms"`
	Root               string          `json:"root,omitempty"                 yaml:"root" xml:"root"`
	Identity           string          `json:"identity,omitempty"             yaml:"identity" xml:"identity"`
}

// ZookeeperNode defines item of nodes section
type ZookeeperNode struct {
	Host string `json:"host" yaml:"host" xml:"host"`
	Port int32  `json:"port" yaml:"port" xml:"port"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClickHouseCluster is the Schema for the clickhouseclusters API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=clickhouseclusters,scope=Namespaced,shortName=chc
type ClickHouseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClickHouseClusterSpec   `json:"spec,omitempty"`
	Status ClickHouseClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClickHouseClusterList contains a list of ClickHouseCluster
type ClickHouseClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClickHouseCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClickHouseCluster{}, &ClickHouseClusterList{})
}
