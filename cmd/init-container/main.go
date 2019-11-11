package main

import "os"

const (
	PodName = "POD_NAME"
)

func main() {
	pod := os.Getenv(PodName)

}
