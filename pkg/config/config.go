package config

import (
	clickhousev1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

const configFile = "config.yaml"

type DefaultConfig struct {
	DefaultClickhouseImage     string                        `json:"default_clickhouse_image"`
	DefaultClickhouseInitImage string                        `json:"default_clickhouse_init_image"`
	DefaultShardCount          int32                         `json:"default_shard_count"`
	DefaultReplicasCount       int32                         `json:"default_replicas_count"`
	DefaultConfig              []string                      `json:"default_config"`
	defaultXMLConfig           map[string]string             `json:"default_xml_config"`
	DefaultZookeeper           *clickhousev1.ZookeeperConfig `json:"default_zookeeper"`
}

func (d *DefaultConfig) GetDefaultXMLConfig() map[string]string {
	return d.defaultXMLConfig
}

func LoadDefaultConfig() (*DefaultConfig, error) {
	config := new(DefaultConfig)
	config.defaultXMLConfig = make(map[string]string)
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, path := range config.DefaultConfig {
		c, err := ioutil.ReadFile(path)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		config.defaultXMLConfig[filepath.Base(path)] = string(c)
	}

	return config, nil
}
