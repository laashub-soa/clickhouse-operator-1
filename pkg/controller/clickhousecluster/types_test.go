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
					}},
			},
		},
	}
	x, err := yandex.ToString()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(x)
}
