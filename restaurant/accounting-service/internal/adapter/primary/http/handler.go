package http

import "github.com/JIeeiroSst/accounting-service/internal/domain/port"

type Handler struct {
	authCart port.AuthCartUsecase
	payment  port.PaymentUsecase
}

func NewHandler(authCart port.AuthCartUsecase, payment port.PaymentUsecase) *Handler {
	return &Handler{authCart: authCart, payment: payment}
}
