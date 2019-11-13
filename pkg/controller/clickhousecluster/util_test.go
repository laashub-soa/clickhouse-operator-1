package clickhousecluster

import (
	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"testing"
)

func TestParseRemoteServersXML(t *testing.T) {
	server := RemoteServers{
		RemoteServer: map[string]Cluster{
			"demo": {
				Shard: []Shard{
					{
						InternalReplication: false,
						Replica: []Replica{
							{
								Host: "host1",
								Port: 8080,
							},
						},
					},
				},
			},
			"another": {
				Shard: []Shard{
					{
						InternalReplication: false,
						Replica: []Replica{
							{
								Host: "host1",
								Port: 8081,
							},
							{
								Host: "host2",
								Port: 8082,
							},
						},
					},
					{
						InternalReplication: true,
						Replica: []Replica{
							{
								Host: "host3",
								Port: 8083,
							},
							{
								Host: "host4",
								Port: 8084,
							},
						},
					},
				},
			},
		},
	}
	t.Log(ParseXML(server))
}

func TestParseZookeeperXML(t *testing.T) {
	zkc := v1.ZookeeperConfig{
		Nodes: []v1.ZookeeperNode{
			{
				Host: "zk1",
				Port: 2181,
			},
			{
				Host: "zk2",
				Port: 2182,
			},
		},
		SessionTimeoutMs:   1000,
		OperationTimeoutMs: 2000,
		Root:               "/zookeeper/test",
		Identity:           "abc",
	}

	zk := Zookeeper{
		Zookeeper: &zkc,
	}
	t.Log(ParseXML(zk))
}
