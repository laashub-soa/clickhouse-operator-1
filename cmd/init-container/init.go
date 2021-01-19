package init_container

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	v1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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

	if zkc.Zookeeper == nil || zkc.Zookeeper.Nodes == nil || zkc.Zookeeper.Nodes[0].Host == "" {
		logrus.Info("no zookeeper configuration")
		return nil
	}

	for _, host := range zkc.Zookeeper.Nodes {
		logrus.Infof("add zk node: %s/%s\n", host.Host, host.Port)
		hosts = append(hosts, fmt.Sprintf("%s:%d", host.Host, host.Port))
	}
	conn, _, err := zk.Connect(hosts, time.Second*10)
	if err != nil {
		return fmt.Errorf("connect zk err: %s", err.Error())
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
			_, err = conn.Create(path, []byte("data"), 0, acls)
			if err == nil {
				logrus.Infof("create path: %s successfully", path)
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

func createMacrosFile() error {
	pod := os.Getenv(PodName)
	if pod == "" {
		return fmt.Errorf("can not find pod name")
	}
	content, err := ioutil.ReadFile(MacrosJSON)
	if err != nil {
		return err
	}
	macros := make(map[string]string)
	err = json.Unmarshal(content, &macros)
	if err != nil {
		return err
	}
	if c, ok := macros[pod]; ok {
		err = ioutil.WriteFile(MacrosXML, []byte(c), 0644)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("can not find %s in %s", pod, content)
}

func Run(_ *cli.Context) error {
	err := createMacrosFile()
	if err != nil {
		logrus.Errorf(err.Error())
		return err
	}
	err = createZookeeperNode()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}
