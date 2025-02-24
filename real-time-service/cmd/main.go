package main

import (
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/real-time-service/config"
	httpCall "github.com/JIeeiroSst/real-time-service/internal/delivery/http"
	"github.com/JIeeiroSst/real-time-service/internal/delivery/middleware"
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

	consul := consul.NewConfigConsul(dirEnv.HostConsul, 
		dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Logger().Error(err.Error())
	}

	wsDelivery := ws.NewWsDelivery(conf)
	httpDelivery := httpCall.NewHttpDelivery(conf)
	middleware := middleware.NewMiddlewareDelivery(conf)

	router := router.NewRouter(wsDelivery, httpDelivery, middleware)
	router.HandlerRouter()

	http.ListenAndServe(fmt.Sprintf(":%v", conf.Server.ServerPort), nil)
	logger.Logger().Sugar().Infof("Server started at port %v", conf.Server.ServerPort)
}
