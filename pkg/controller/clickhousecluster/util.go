package clickhousecluster

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"reflect"
	"strings"
)

func doParse(v reflect.Value, indent int, father string) string {
	var out string
	space := strings.Repeat("   ", indent)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map {
		for _, e := range v.MapKeys() {
			tag := e.String()
			out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, tag, doParse(v.MapIndex(e), indent+1, tag), space, e)
		}
	} else if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			tag := v.Type().Field(i).Tag.Get("xml")
			if tag != "" {
				kind := v.Field(i).Kind()
				if kind == reflect.Slice {
					out += fmt.Sprintf("%s", doParse(v.Field(i), indent, tag))
					continue
				}
				if kind == reflect.String || kind == reflect.Bool || kind == reflect.Int || kind == reflect.Int32 {
					out += fmt.Sprintf("\n%s<%s>%s</%s>", space, tag, doParse(v.Field(i), indent+1, tag), tag)
					continue
				}
				out += fmt.Sprintf("%s<%s>%s\n%s</%s>", space, tag, doParse(v.Field(i), indent+1, tag), space, tag)
			}
		}
	} else if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, father, doParse(v.Index(i), indent+1, ""), space, father)
		}
	} else {
		out = fmt.Sprintf("%v", v)
	}
	return out
}

func ParseXML(s interface{}) string {
	v := reflect.ValueOf(s)
	return fmt.Sprintf("<yandex>\n%s\n</yandex>", doParse(v, 1, ""))
}

func isStatefulSetReady(statefulSet *appsv1.StatefulSet) bool {
	return statefulSet.Status.Replicas == *statefulSet.Spec.Replicas &&
		statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas
}
