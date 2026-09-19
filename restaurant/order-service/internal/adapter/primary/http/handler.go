package http

import "github.com/JIeeiroSst/order-service/internal/domain/port"

type Handler struct {
	order       port.OrderUsecase
	reservation port.ReservationUsecase
}

func NewHandler(order port.OrderUsecase, reservation port.ReservationUsecase) *Handler {
	return &Handler{order: order, reservation: reservation}
}
