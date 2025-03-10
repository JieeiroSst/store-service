package v1

import (
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/accounting-service/common"
	"github.com/JIeeiroSst/accounting-service/internal/dto"
	"github.com/JIeeiroSst/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initAuthCartRoutes(api *gin.RouterGroup) {
	group := api.Group("/accounting")

	group.POST("", h.AuthCart)
}

func (h *Handler) AuthCart(ctx *gin.Context) {
	now := time.Now()
	hour := now.Hour()

	var authCart dto.AuthCart
	if err := ctx.ShouldBindJSON(&authCart); err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Error:   false,
			Message: err.Error(),
		})
	}
	orderJson, err := json.Marshal(&authCart.Order)
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Error:   false,
			Message: err.Error(),
		})
	}

	deliveryJson, err := json.Marshal(&authCart.Delivery)
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Error:   false,
			Message: err.Error(),
		})
	}

	if hour > 8 && hour < 23 {
		err := h.nats.Publish(common.OrderSuccess, orderJson)
		if err != nil {
			response.ResponseStatus(ctx, 500, response.MessageStatus{
				Error:   false,
				Message: err.Error(),
			})
		} else {
			err := h.nats.Publish(common.DeliveryShip, deliveryJson)
			if err != nil {
				response.ResponseStatus(ctx, 500, response.MessageStatus{
					Error:   false,
					Message: err.Error(),
				})
			} else {
				if err := h.usecase.SaveDelivery(ctx, authCart.Delivery); err != nil {
					response.ResponseStatus(ctx, 500, response.MessageStatus{
						Error:   false,
						Message: err.Error(),
					})
				}
				if err := h.usecase.SaveOrder(ctx, authCart.Order, common.OrderSuccess); err != nil {
					response.ResponseStatus(ctx, 500, response.MessageStatus{
						Error:   false,
						Message: err.Error(),
					})
				}
			}
		}
	} else {
		if err := h.nats.Publish(common.OrderReject, orderJson); err != nil {
			response.ResponseStatus(ctx, 500, response.MessageStatus{
				Error:   false,
				Message: err.Error(),
			})
		} else {
			if err := h.usecase.SaveOrder(ctx, authCart.Order, common.OrderReject); err != nil {
				response.ResponseStatus(ctx, 500, response.MessageStatus{
					Error:   false,
					Message: err.Error(),
				})
			}
		}
	}

	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: "success",
	})
}
