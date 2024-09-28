package v1

import (
	"encoding/json"

	"github.com/JIeeiroSst/consumer-service/internal/dto"
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initConsumerRoutes(api *gin.RouterGroup) {
	group := api.Group("/consumer")

	group.GET("", h.FindConsumer)
	group.POST("", h.ConsumerOrder)
}

func (h *Handler) FindConsumer(ctx *gin.Context) {
	var pagination logger.Pagination
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}
	consumer, err := h.usecase.Consumers.Find(ctx, pagination)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"data": consumer,
	})
}

func (h *Handler) ConsumerOrder(ctx *gin.Context) {
	var consumer dto.Consumer
	if err := ctx.ShouldBindJSON(&consumer); err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}

	consumerJson, err := json.Marshal(&consumer)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	if err := h.nats.Publish("kitchen.create", consumerJson); err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	order := consumer.BuildV2()
	orderJson, err := json.Marshal(&order)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	if err := h.nats.Publish("order.created", orderJson); err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
	})
}
