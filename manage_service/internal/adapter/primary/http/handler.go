package http

import "github.com/JIeeiroSst/manage-service/internal/domain/port"

type Handler struct {
	auth     port.AuthUsecase
	keycloak port.KeycloakUsecase
}

func NewHandler(auth port.AuthUsecase, keycloak port.KeycloakUsecase) *Handler {
	return &Handler{
		auth:     auth,
		keycloak: keycloak,
	}
}
