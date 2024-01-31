package http

import "github.com/JIeeiroSst/bookStore-service/middleware"

type Handler struct {
	middleware middleware.Middleware
}

func NewHandler(middleware middleware.Middleware) *Handler {
	return &Handler{
		middleware: middleware,
	}
}
