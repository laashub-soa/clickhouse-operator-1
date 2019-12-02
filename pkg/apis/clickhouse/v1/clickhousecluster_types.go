package v1

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationLastApplied string = "clickhouse.service.diamond.sensetime.com/last-applied-configuration"
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

	Pod *PodPolicy `json:"pod,omitempty"`

	// Pod defines the policy for pods owned by clickhouse operator.
	// This field cannot be updated once the CR is created.
	Resources ClickHouseResources `json:"resources,omitempty"`
}

// Remove subresources, cuz https://github.com/kubernetes/kubectl/issues/564
// ClickHouseClusterStatus defines the observed state of ClickHouseCluster
// +k8s:openapi-gen=true
type ClickHouseClusterStatus struct {
	Phase       string                  `json:"phase,omitempty"`
	ShardStatus map[string]*ShardStatus `json:"shardStatus,omitempty"`
}

// ClickHouseResources sets the limits and requests for a container
type ClickHouseResources struct {
	Requests CPUAndMem `json:"requests,omitempty"`
	Limits   CPUAndMem `json:"limits,omitempty"`
}

// CPUAndMem defines how many cpu and ram the container will request/limit
type CPUAndMem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

//ShardStatus defines states of Clickhouse for 1 shard (1 statefulset)
type ShardStatus struct {
	// Phase indicates the state this Clickhouse cluster jumps in.
	// Phase goes as one way as below:
	//   Initial -> Running <-> updating
	Phase string `json:"phase,omitempty"`
}

// PodPolicy defines the policy for pods owned by ClickHouse operator.
type PodPolicy struct {
	// Annotations specifies the annotations to attach to headless service the ClickHouse operator creates
	Annotations map[string]string `json:"annotations,omitempty"`
	// Tolerations specifies the tolerations to attach to the pods the ClickHouse operator creates
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
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

type CustomPodSpec struct {
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

func (cc *ClickHouseCluster) ComputeLastAppliedConfiguration() (string, error) {
	lastcc := cc.DeepCopy()
	//remove unnecessary fields
	lastcc.Annotations = nil
	lastcc.ResourceVersion = ""
	lastcc.Status = ClickHouseClusterStatus{}

	lastApplied, err := json.Marshal(lastcc)
	if err != nil {
		logrus.Errorf("[%s]: Cannot create last-applied-configuration = %v", cc.Name, err)
	}
	return string(lastApplied), err
}

func init() {
	SchemeBuilder.Register(&ClickHouseCluster{}, &ClickHouseClusterList{})
}
