package clickhousecluster

import (
	"fmt"
	"reflect"
	"strings"
)

func doParse(v reflect.Value, indent int) string {
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
	} else if t.Kind() == reflect.Slice {
		fmt.Println(t)
		for i := 0; i < v.Len(); i++ {
			out += doParse(v.Index(i), indent+1)
		}
	} else {
		out = fmt.Sprintf("\n%s%v", space, v)
	}
	return out
}

func ParseXML(s interface{}) string {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return fmt.Sprintf("%s", doParse(v, 1))
}
