package main

import (
	"log"
	"net/http"

	"github.com/JIeeiroSst/real-time-service/config"
	"github.com/JIeeiroSst/real-time-service/internal/delivery"
	httpCall "github.com/JIeeiroSst/real-time-service/internal/delivery/http"
	"github.com/JIeeiroSst/real-time-service/internal/delivery/ws"
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

	http.Handle("/ws", delivery.Middleware(
		http.HandlerFunc(wsDelivery.WsHandler),
		delivery.AuthMiddleware,
	))

	http.Handle("/call", delivery.Middleware(
		http.HandlerFunc(httpDelivery.WsCall),
		delivery.AuthMiddleware,
	))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
