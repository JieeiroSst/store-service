package main

import (
	"log"
	"os"
	"strings"

	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/consul"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	nodeEnv := os.Getenv("production")

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Println(err)
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Println(err)
		}
	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Println(err)
		}
	}
}
