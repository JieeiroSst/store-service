package http

import "github.com/JIeeiroSst/delivery-service/internal/domain/port"

type Handler struct {
	delivery port.DeliveryUsecase
}

func NewHandler(delivery port.DeliveryUsecase) *Handler {
	return &Handler{delivery: delivery}
}
