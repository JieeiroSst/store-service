package http

import "github.com/gofiber/fiber/v2"

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Init(router fiber.App) {
	h.corsMiddleware(router)
	h.initApi(router)
}

func (h *Handler) initApi(router fiber.App) {

}
