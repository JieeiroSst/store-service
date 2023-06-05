package main

import "github.com/JIeeiroSst/workflow-service/pkg/log"

func main() {
	err := log.Init("info", "stdout")
	if err != nil {
		panic(err)
	}
}
