package clickhousecluster

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

type YandexRemoteServers struct {
	Yandex RemoteServers `xml:"yandex"`
}
