package ws

import (
	"net/http"

	"github.com/JIeeiroSst/real-time-service/config"
	"github.com/JIeeiroSst/real-time-service/pkg/logger"
	"github.com/gorilla/websocket"
)

type WsDelivery struct {
	config *config.Config
}

func NewWsDelivery(config *config.Config) *WsDelivery {
	return &WsDelivery{
		config: config,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ww *WsDelivery) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger().Error(err.Error())
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		logger.Logger().Sugar().Infof("Received message:", string(message))

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			logger.Logger().Error(err.Error())
			return
		}
	}
}
