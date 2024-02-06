package router

import (
	"net/http"

	httpCall "github.com/JIeeiroSst/real-time-service/internal/delivery/http"
	"github.com/JIeeiroSst/real-time-service/internal/delivery/middleware"
	"github.com/JIeeiroSst/real-time-service/internal/delivery/ws"
)

type Router struct {
	wsDelivery   *ws.WsDelivery
	httpDelivery *httpCall.HttpDelivery
	middleware   *middleware.MiddlewareDelivery
}

func NewRouter(wsDelivery *ws.WsDelivery,
	httpDelivery *httpCall.HttpDelivery,
	middleware *middleware.MiddlewareDelivery) *Router {
	return &Router{
		wsDelivery:   wsDelivery,
		httpDelivery: httpDelivery,
		middleware:   middleware,
	}
}

func (r *Router) HandlerRouter() {
	http.Handle("/ws", middleware.Middleware(
		http.HandlerFunc(r.wsDelivery.WsHandler),
		r.middleware.AuthMiddleware,
	))

	http.Handle("/call", middleware.Middleware(
		http.HandlerFunc(r.httpDelivery.WsCall),
		r.middleware.AuthMiddleware,
	))
}
