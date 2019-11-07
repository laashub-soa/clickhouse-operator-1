package clickhousecluster

import (
	"fmt"
	"reflect"
	"strings"
)

func doParse(v reflect.Value, indent int, father string) string {
	var out string
	space := strings.Repeat("\t", indent)
	t := v.Type()
	if t.Kind() == reflect.Map {
		for _, e := range v.MapKeys() {
			tag := e.String()
			out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, tag, doParse(v.MapIndex(e), indent+1, tag), space, e)
		}
	} else if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("xml")
			if tag != "" {
				if v.Field(i).Kind() == reflect.Slice {
					out += fmt.Sprintf("%s", doParse(v.Field(i), indent+1, tag))
					continue
				}
				if v.Field(i).Kind() == reflect.String || v.Field(i).Kind() == reflect.Bool || v.Field(i).Kind() == reflect.Int {
					out += fmt.Sprintf("\n%s<%s>%s</%s>", space, tag, doParse(v.Field(i), indent+1, tag), tag)
					continue
				}
				out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, tag, doParse(v.Field(i), indent+1, tag), space, tag)
			}
		}
	} else if t.Kind() == reflect.Slice {
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
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return fmt.Sprintf("%s", doParse(v, 1, "yandex"))
}
