package main

import (
	"log"
	"net/http"

	"github.com/JIeeiroSst/real-time-service/internal/delivery"
	httpCall "github.com/JIeeiroSst/real-time-service/internal/delivery/http"
	"github.com/JIeeiroSst/real-time-service/internal/delivery/ws"
)

func main() {
	http.Handle("/ws", delivery.Middleware(
		http.HandlerFunc(ws.WsHandler),
		delivery.AuthMiddleware,
	))

	http.Handle("/call", delivery.Middleware(
		http.HandlerFunc(httpCall.WsCall),
		delivery.AuthMiddleware,
	))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
