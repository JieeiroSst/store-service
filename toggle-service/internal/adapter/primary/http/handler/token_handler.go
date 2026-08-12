package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/dto"
	adminmw "github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type TokenHandler struct {
	tokens port.TokenService
}

func NewTokenHandler(tokens port.TokenService) *TokenHandler {
	return &TokenHandler{tokens: tokens}
}

func (h *TokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTokenRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}

	in := port.CreateTokenInput{Name: req.Name, Type: model.APITokenType(req.Type)}
	if req.ProjectID != nil {
		id, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			writeError(w, apperr.ErrValidation)
			return
		}
		in.ProjectID = &id
	}
	if req.EnvironmentID != nil {
		id, err := uuid.Parse(*req.EnvironmentID)
		if err != nil {
			writeError(w, apperr.ErrValidation)
			return
		}
		in.EnvironmentID = &id
	}
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			writeError(w, apperr.ErrValidation)
			return
		}
		in.ExpiresAt = &t
	}

	userID, _ := adminmw.UserID(r.Context())
	plaintext, tok, err := h.tokens.Create(r.Context(), in, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": plaintext, "record": tok})
}

func (h *TokenHandler) List(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.tokens.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (h *TokenHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "tokenId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	if err := h.tokens.Revoke(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
