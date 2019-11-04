package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationLastApplied string = "clickhouse.sensetime.com/last-applied-configuration"

	ClusterPhaseInitial string = "Initializing"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClickHouseClusterSpec defines the desired state of ClickHouseCluster
// +k8s:openapi-gen=true
type ClickHouseClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Cluster      Cluster `json:"clusters,omitempty"`
	DataCapacity string  `json:"dataCapacity,omitempty"`

	//Define StorageClass for Persistent Volume Claims in the local storage.
	DataStorageClass string `json:"dataStorageClass,omitempty"`
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
	if c.Spec.Cluster.Name == "" {
		c.Spec.Cluster.Name = c.Name
		changed = true
	}
	if c.Spec.Cluster.ShardsCount == 0 {
		c.Spec.Cluster.ShardsCount = 1
		changed = true
	}
	if c.Spec.Cluster.ReplicasCount == 0 {
		c.Spec.Cluster.ReplicasCount = 1
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

type Cluster struct {
	Name          string `json:"name"`
	ShardsCount   int    `json:"shardsCount,omitempty"`
	ReplicasCount int    `json:"replicasCount,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ClickHouseCluster{}, &ClickHouseClusterList{})
}
