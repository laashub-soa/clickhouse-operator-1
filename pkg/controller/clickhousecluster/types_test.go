package clickhousecluster

import (
	"testing"
)

func TestParseXML(t *testing.T) {
	yandex := YandexRemoteServers{
		Yandex: RemoteServers{
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
		},
	}
	t.Log(ParseXML(yandex))
}
