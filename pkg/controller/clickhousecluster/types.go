package clickhousecluster

import (
	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
)

const (
	ClusterPhaseInitial  = "Initializing"
	ClusterPhaseCreating = "Creating"
	ClusterPhaseUpdating = "Updating"
	ClusterPhaseRunning  = "Running"

	ShardPhaseRunning = "Running"
	ShardPhaseInitial = "Initializing"

	ShardIDLabelKey  = "shard-id"
	CreateByLabelKey = "created-by"
	ClusterLabelKey  = "clickhouse-cluster"

	OperatorLabelKey = "clickhouse-operator"
)

type Replica struct {
	Host     string `xml:"host"`
	Port     int    `xml:"port"`
	Password string `xml:"password"`
	User     string `xml:"user"`
}

type Shard struct {
	InternalReplication bool      `xml:"internal_replication"`
	Replica             []Replica `xml:"replica"`
}

type Cluster struct {
	Shard []Shard `xml:"shard"`
}

type RemoteServers struct {
	RemoteServer map[string]Cluster `xml:"remote_servers"`
}

type Zookeeper struct {
	Zookeeper *v1.ZookeeperConfig `xml:"zookeeper"`
}
