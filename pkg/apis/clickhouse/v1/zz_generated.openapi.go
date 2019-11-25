// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ClickHouseCluster":       schema_pkg_apis_clickhouse_v1_ClickHouseCluster(ref),
		"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ClickHouseClusterSpec":   schema_pkg_apis_clickhouse_v1_ClickHouseClusterSpec(ref),
		"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ClickHouseClusterStatus": schema_pkg_apis_clickhouse_v1_ClickHouseClusterStatus(ref),
	}
}

func schema_pkg_apis_clickhouse_v1_ClickHouseCluster(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ClickHouseCluster is the Schema for the clickhouseclusters API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ClickHouseClusterSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ClickHouseClusterStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ClickHouseClusterSpec", "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ClickHouseClusterStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_clickhouse_v1_ClickHouseClusterSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ClickHouseClusterSpec defines the desired state of ClickHouseCluster",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"image": {
						SchemaProps: spec.SchemaProps{
							Description: "ClickHouse Docker image",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"initImage": {
						SchemaProps: spec.SchemaProps{
							Description: "ClickHouse init  image",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"deletePVC": {
						SchemaProps: spec.SchemaProps{
							Description: "DeletePVC defines if the PVC must be deleted when the cluster is deleted it is false by default",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"shardsCount": {
						SchemaProps: spec.SchemaProps{
							Description: "Shards count",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"replicasCount": {
						SchemaProps: spec.SchemaProps{
							Description: "Replicas count",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"zookeeper": {
						SchemaProps: spec.SchemaProps{
							Description: "Zookeeper config",
							Ref:         ref("github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ZookeeperConfig"),
						},
					},
					"CustomSettings": {
						SchemaProps: spec.SchemaProps{
							Description: "Custom defined XML settings, like <yandex>somethine</yandex>",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"dataCapacity": {
						SchemaProps: spec.SchemaProps{
							Description: "The storage capacity",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"dataStorageClass": {
						SchemaProps: spec.SchemaProps{
							Description: "Define StorageClass for Persistent Volume Claims in the local storage.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"pod": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.PodPolicy"),
						},
					},
				},
				Required: []string{"CustomSettings"},
			},
		},
		Dependencies: []string{
			"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.PodPolicy", "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ZookeeperConfig"},
	}
}

func schema_pkg_apis_clickhouse_v1_ClickHouseClusterStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Remove subresources, cuz https://github.com/kubernetes/kubectl/issues/564 ClickHouseClusterStatus defines the observed state of ClickHouseCluster",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"phase": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"shardStatus": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ShardStatus"),
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1.ShardStatus"},
	}
}
