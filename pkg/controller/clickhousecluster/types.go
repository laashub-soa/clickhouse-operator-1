package clickhousecluster

import (
	"fmt"
	"reflect"
	"strings"
)

type Replica struct {
	Host string `xml:"host"`
	Port string `xml:"port"`
}

type Shard struct {
	InternalReplication bool    `xml:"internal_replication"`
	Replica             Replica `xml:"replica"`
}

type Cluster struct {
	Shard Shard `xml:"shard"`
}

type RemoteServers struct {
	RemoteServer map[string]Cluster `xml:"remote_servers"`
}

type YandexRemoteServers struct {
	Yandex RemoteServers `xml:"yandex"`
}

func doParse(v reflect.Value, indent int) interface{} {
	var out string
	space := strings.Repeat("\t", indent)
	t := v.Type()
	if t.Kind() == reflect.Map {
		for _, e := range v.MapKeys() {
			out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, e, doParse(v.MapIndex(e), indent+1), space, e)
		}
	} else if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("xml")
			if tag != "" {
				out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, tag, doParse(v.Field(i), indent+1), space, tag)
			}
		}
	} else {
		return fmt.Sprintf("\n%s%v", space, v)
	}
	return out
}

func ParseXML(s interface{}) string {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>%s`, doParse(v, 1))
}
