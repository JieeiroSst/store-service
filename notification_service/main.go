package main

import (
	"log"

	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/consul"
)

func main() {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		log.Println(err)
	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		log.Println(err)
	}

}
