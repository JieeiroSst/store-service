package v1

import (
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/accounting-service/internal/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initAuthCartRoutes(api *gin.RouterGroup) {
	group := api.Group("/accounting")

	group.POST("", h.AuthCart)
}

func (h *Handler) AuthCart(ctx *gin.Context) {
	now := time.Now()
	hour := now.Hour()

	var order dto.Order
	if err := ctx.ShouldBindJSON(&order); err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}
	orderJson, err := json.Marshal(&order)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	if hour > 8 && hour < 23 {
		if err := h.nats.Publish("order.success", orderJson); err != nil {
			ctx.JSON(500, gin.H{
				"error": err,
			})
			return
		}
	} else {
		if err := h.nats.Publish("order.reject", orderJson); err != nil {
			ctx.JSON(500, gin.H{
				"error": err,
			})
			return
		}
	}

	ctx.JSON(200, gin.H{
		"sucess": true,
	})
}
