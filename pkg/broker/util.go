package broker

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func toJson(data interface{}) string {
	o, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(o)
}

func BytesToStringSlice(in []byte) []string {
	var (
		findRight bool
		left      int
		out       = make([]string, 0)
	)
	s := string(in)
	for i := 0; i < len(s); i++ {
		w := string(s[i])
		if w == "[" || w == "]" {
			continue
		}
		if w == "\"" {
			if findRight {
				out = append(out, s[left+1:i])
			}
			left = i
			findRight = !findRight
		}
	}
	return out
}
