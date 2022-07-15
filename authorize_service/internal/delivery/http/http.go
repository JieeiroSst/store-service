package http

import (
	v1 "github.com/JieeiroSst/authorize-service/internal/delivery/http/v1"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase, adapter persist.Adapter) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.initApi(router)
	h.corsMiddleware(router)
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(&h.usecase)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
