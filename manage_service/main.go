package main

import (
	"embed"
	"os"
	"strings"

	"github.com/JIeeiroSst/manage-service/config"
	"github.com/JIeeiroSst/manage-service/pkg/consul"
	"github.com/gin-gonic/gin"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

func main() {
	router := gin.Default()
	nodeEnv := os.Getenv("NODE_ENV")

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
		}
	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
		}
	}

	// app := app.NewApp(conf)

	// go func() {
	// 	app.NewGRPCServer()
	// }()

	// app.NewServerApp(router)

	router.Run()
}
