package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/samuel/go-zookeeper/zk"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	PodName      = "POD_NAME"
	MacrosJSON   = "/etc/clickhouse-server/config.d/all-macros.json"
	MacrosXML    = "/etc/clickhouse-server/conf.d/macros.xml"
	ZookeeperXML = "/etc/clickhouse-server/config.d/zookeeper.xml"
)

type ZkConfig struct {
	Zookeeper *v1.ZookeeperConfig `xml:"zookeeper"`
}

func createZookeeperNode() error {
	var (
		zkc   ZkConfig
		acls  = zk.WorldACL(zk.PermAll)
		hosts = make([]string, 0)
	)

	context, err := ioutil.ReadFile(ZookeeperXML)
	if err != nil {
		return err
	}

	if err := xml.Unmarshal(context, &zkc); err != nil {
		return err
	}
	for _, host := range zkc.Zookeeper.Nodes {
		hosts = append(hosts, fmt.Sprintf("%s:%d", host.Host, host.Port))
	}
	conn, _, err := zk.Connect(hosts, time.Second*10)
	if err != nil {
		return err
	}
	if zkc.Zookeeper.Identity != "" {
		err = conn.AddAuth("digest", []byte(zkc.Zookeeper.Identity))
		if err != nil {
			return err
		}
	}
	defer conn.Close()

	create := func(path string) error {
		exist, _, err := conn.Exists(path)
		if err != nil {
			return err
		}
		if !exist {
			_, err = conn.Create(path, []byte("hello"), 0, acls)
			if err == nil {
				log.Printf("create path: %s successfully", path)
			}
			return err
		}
		return nil
	}

	return func(root string, f func(string) error) error {
		if !strings.HasPrefix(root, "/") {
			root = "/" + root
		}
		dirs := strings.Split(root, "/")
		for i := 1; i < len(dirs)+1; i++ {
			path := strings.Join(dirs[:i], "/")
			if path != "" {
				if err := f(path); err != nil {
					return err
				}
			}
		}
		return nil
	}(zkc.Zookeeper.Root, create)
}

func createMarosFile() error {
	pod := os.Getenv(PodName)
	if pod == "" {
		return fmt.Errorf("can not find pod name")
	}
	content, err := ioutil.ReadFile(MacrosJSON)
	if err != nil {
		return err
	}
	maros := make(map[string]string)
	err = json.Unmarshal(content, &maros)
	if err != nil {
		return err
	}
	if c, ok := maros[pod]; ok {
		err = ioutil.WriteFile(MacrosXML, []byte(c), 0644)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("can not find %s in %s", pod, content)
}

func main() {
	err := createMarosFile()
	if err != nil {
		log.Fatal(err)
	}
	err = createZookeeperNode()
	if err != nil {
		log.Fatal(err)
	}
}
