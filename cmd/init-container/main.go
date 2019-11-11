package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	PodName    = "POD_NAME"
	MacrosJSON = "/etc/clickhouse-server/all-macros.json"
	MacrosXML  = "/etc/clickhouse-server/conf.d/macros.xml"
)

func main() {
	pod := os.Getenv(PodName)
	if pod == "" {
		fmt.Println("can not find pod name")
		os.Exit(-1)
	}
	content, err := ioutil.ReadFile(MacrosJSON)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	maros := make(map[string]string)
	err = json.Unmarshal(content, maros)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if c, ok := maros[pod]; ok {
		err = ioutil.WriteFile(MacrosXML, []byte(c), 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}
	fmt.Printf("can not find %s in %s", pod, content)
	os.Exit(-1)

}
