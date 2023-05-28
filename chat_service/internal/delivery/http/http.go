package http

import (
	"github.com/JIeeiroSst/chat-service/config"
	v1 "github.com/JIeeiroSst/chat-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/chat-service/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type Http struct {
	Usecase *usecase.Usecase
	Config  *config.Config
}

func NewHttp(Usecase *usecase.Usecase, Config *config.Config) *Http {
	return &Http{
		Usecase: Usecase,
	}
}

func (h *Http) Init(router chi.Router) {
	h.corsMiddleware(router)
	h.initApi(router)
}

func (h *Http) initApi(router chi.Router) {
	handlerV1 := v1.NewHttpV1(h.Usecase)
	handlerV1.SetupRoutes(router)
}
