package http

import (
	v1 "github.com/JIeeiroSst/upload-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/upload-service/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) Init(router *fiber.App) {
	h.corsMiddleware(router)
	h.initApi(router)
}

func (h *Handler) initApi(router *fiber.App) {
	handlerV1 := v1.NewHandler(h.usecase)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
