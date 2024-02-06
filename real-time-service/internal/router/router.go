package router

import (
	"net/http"

	"github.com/JIeeiroSst/real-time-service/internal/delivery"
	httpCall "github.com/JIeeiroSst/real-time-service/internal/delivery/http"
	"github.com/JIeeiroSst/real-time-service/internal/delivery/ws"
)

type Router struct {
	wsDelivery   *ws.WsDelivery
	httpDelivery *httpCall.HttpDelivery
}

func NewRouter(wsDelivery *ws.WsDelivery,
	httpDelivery *httpCall.HttpDelivery) *Router {
	return &Router{
		wsDelivery:   wsDelivery,
		httpDelivery: httpDelivery,
	}
}

func (r *Router) HandlerRouter() {
	http.Handle("/ws", delivery.Middleware(
		http.HandlerFunc(r.wsDelivery.WsHandler),
		delivery.AuthMiddleware,
	))

	http.Handle("/call", delivery.Middleware(
		http.HandlerFunc(r.httpDelivery.WsCall),
		delivery.AuthMiddleware,
	))
}
