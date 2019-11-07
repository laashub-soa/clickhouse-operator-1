package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationLastApplied string = "clickhouse.sensetime.com/last-applied-configuration"
	ClusterPhaseInitial   string = "Initializing"

	DefaultClickHouseImage string = "registry.sensetime.com/diamond/clickhouse-server:latest"
)

// ClickHouseClusterSpec defines the desired state of ClickHouseCluster
// +k8s:openapi-gen=true
type ClickHouseClusterSpec struct {
	//ClickHouse Docker image
	Image string `json:"image,omitempty"`

	//DeletePVC defines if the PVC must be deleted when the cluster is deleted
	//it is false by default
	DeletePVC bool `json:"deletePVC,omitempty"`

	//Shards count
	ShardsCount int32 `json:"shardsCount,omitempty"`

	//Replicas count
	ReplicasCount int32 `json:"replicasCount,omitempty"`

	//The storage capacity
	DataCapacity string `json:"dataCapacity,omitempty"`

	//Define StorageClass for Persistent Volume Claims in the local storage.
	DataStorageClass string `json:"dataStorageClass,omitempty"`

	//User defined pod spec
	PodSpec *corev1.PodSpec `json:"podSpec,omitempty"`
}

// ClickHouseClusterStatus defines the observed state of ClickHouseCluster
// +k8s:openapi-gen=true
type ClickHouseClusterStatus struct {
	Status string `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClickHouseCluster is the Schema for the clickhouseclusters API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clickhouseclusters,scope=Namespaced
type ClickHouseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClickHouseClusterSpec   `json:"spec,omitempty"`
	Status ClickHouseClusterStatus `json:"status,omitempty"`
}

func (c *ClickHouseCluster) SetDefaults() bool {
	var changed = false
	if c.Status.Status == "" {
		c.Status.Status = ClusterPhaseInitial
		changed = true
	}
	if c.Spec.Image == "" {
		c.Spec.Image = DefaultClickHouseImage
		changed = true
	}
	if c.Spec.ShardsCount == 0 {
		c.Spec.ShardsCount = 1
		changed = true
	}
	if c.Spec.ReplicasCount == 0 {
		c.Spec.ReplicasCount = 1
		changed = true
	}
	return changed
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
