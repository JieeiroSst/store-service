package handler

import (
	"net/http"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/dto"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type AuthHandler struct {
	auth port.AuthService
}

func NewAuthHandler(auth port.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.auth.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	token, u, err := h.auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}
