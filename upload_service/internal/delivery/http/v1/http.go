package v1

import (
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

func (h *Handler) Init(api fiber.Router) {
	v1 := api.Group("/v1")
	{
		h.initUploadRoutes(v1)
	}
}
