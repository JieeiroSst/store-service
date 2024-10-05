package v1

import (
	"github.com/JIeeiroSst/delivery-service/internal/dto"
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initDeliveryRoutes(api *gin.RouterGroup) {
	g := api.Group("/delivery")

	g.POST("/", h.RegisterDelivery)
	g.GET("/", h.All)
}

func (h *Handler) RegisterDelivery(ctx *gin.Context) {
	var delivery dto.Delivery
	if err := ctx.ShouldBind(&delivery); err != nil {
		logger.ResponseStatus(ctx, 400, err)
	}

	if err := h.usecase.Deliveries.Create(ctx, delivery); err != nil {
		logger.ResponseStatus(ctx, 500, err)
	}
}

func (h *Handler) All(ctx *gin.Context) {
	var pagination logger.Pagination
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		logger.ResponseStatus(ctx, 400, err)
	}

	deliveries, err := h.usecase.FindAll(ctx, pagination)
	if err != nil {
		logger.ResponseStatus(ctx, 500, err)
	}

	logger.ResponseStatus(ctx, 200, deliveries)
}
