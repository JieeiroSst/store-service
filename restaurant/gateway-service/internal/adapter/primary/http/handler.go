package http

import (
	"github.com/JIeeiroSst/gateway-service/internal/domain/port"
	"github.com/JIeeiroSst/gateway-service/pkg/authn"
)

type Handler struct {
	gateway port.GatewayUsecase
	auth    *authn.Authenticator
}

func NewHandler(gateway port.GatewayUsecase, auth *authn.Authenticator) *Handler {
	return &Handler{gateway: gateway, auth: auth}
}
