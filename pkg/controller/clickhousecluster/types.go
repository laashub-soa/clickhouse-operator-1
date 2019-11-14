package clickhousecluster

import (
	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
)

const (
	ClusterPhaseInitial  = "Initializing"
	ClusterPhaseUpdating = "Updating"
	ClusterPhaseRunning  = "Running"

	ShardPhaseRunning  = "Running"
	ShardPhaseUpdating = "Updating"
)

type Replica struct {
	Host string `xml:"host"`
	Port int    `xml:"port"`
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
