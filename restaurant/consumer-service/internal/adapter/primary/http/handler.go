package http

import "github.com/JIeeiroSst/consumer-service/internal/domain/port"

type Handler struct {
	consumer port.ConsumerUsecase
}

func NewHandler(consumer port.ConsumerUsecase) *Handler {
	return &Handler{consumer: consumer}
}
