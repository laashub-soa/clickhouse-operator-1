package config

import (
	"fmt"
	clickhousev1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"reflect"
)

const configFile = "/etc/clickhouse-operator/config.yaml"

//const configFile = "./tests/config/config.yaml"

type DefaultConfig struct {
	DefaultClickhouseImage         string                        `yaml:"default_clickhouse_image"`
	DefaultClickhouseInitImage     string                        `yaml:"default_clickhouse_init_image"`
	DefaultClickhouseExporterImage string                        `yaml:"default_clickhouse_exporter_image"`
	DefaultShardCount              int32                         `yaml:"default_shard_count"`
	DefaultReplicasCount           int32                         `yaml:"default_replicas_count"`
	DefaultConfig                  []string                      `yaml:"default_config"`
	defaultXMLConfig               map[string]string             `yaml:"default_xml_config"`
	DefaultZookeeper               *clickhousev1.ZookeeperConfig `yaml:"default_zookeeper"`
	DefaultDataCapacity            string                        `yaml:"default_data_capacity"`
}

func (d *DefaultConfig) GetDefaultXMLConfig() map[string]string {
	return d.defaultXMLConfig
}

func (d *DefaultConfig) validate() error {
	v := reflect.ValueOf(*d)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		//https://golang.org/doc/go1#equality
		if f.Kind() == reflect.Slice || f.Kind() == reflect.Struct || f.Kind() == reflect.Map {
			if reflect.DeepEqual(f, reflect.Zero(f.Type())) {
				return fmt.Errorf("%s is null", v.Type().Field(i).Name)
			}
			continue
		}
		if f.Interface() == reflect.Zero(f.Type()).Interface() {
			return fmt.Errorf("%s is null", v.Type().Field(i).Name)
		}
	}
	return nil
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
	return config, config.validate()
}
