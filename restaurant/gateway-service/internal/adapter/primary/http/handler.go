package http

import "github.com/JIeeiroSst/gateway-service/internal/domain/port"

type Handler struct {
	gateway port.GatewayUsecase
}

func NewHandler(gateway port.GatewayUsecase) *Handler {
	return &Handler{gateway: gateway}
}
