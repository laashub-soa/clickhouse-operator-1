package clickhousecluster

import (
	"testing"
)

func TestToString(t *testing.T) {
	yandex := YandexRemoteServers{
		Yandex: RemoteServers{
			RemoteServer: map[string]Cluster{
				"demo": {
					Shard: Shard{
						InternalReplication: false,
						Replica: Replica{
							Host: "host1",
							Port: "8080",
						},
					},
				},
				"another": {
					Shard: Shard{
						InternalReplication: false,
						Replica: Replica{
							Host: "host2",
							Port: "8081",
						},
					},
				},
			},
		},
	}
	t.Log(ParseXML(yandex))
}
