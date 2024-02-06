package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JIeeiroSst/real-time-service/config"
	httpCall "github.com/JIeeiroSst/real-time-service/internal/delivery/http"
	"github.com/JIeeiroSst/real-time-service/internal/delivery/ws"
	"github.com/JIeeiroSst/real-time-service/internal/router"
	"github.com/JIeeiroSst/real-time-service/pkg/consul"
	"github.com/JIeeiroSst/real-time-service/pkg/logger"
)

func main() {
	nameEnv := ".env"
	dirEnv, err := config.ReadFileEnv(nameEnv)
	if err != nil {
		logger.Logger().Error(err.Error())
	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Logger().Error(err.Error())
	}

	wsDelivery := ws.NewWsDelivery(conf)
	httpDelivery := httpCall.NewHttpDelivery(conf)

	router := router.NewRouter(wsDelivery, httpDelivery)
	router.HandlerRouter()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", conf.Server.ServerPort), nil))
}
