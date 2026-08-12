package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/dto"
	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type EnvironmentHandler struct {
	environments port.EnvironmentService
}

func NewEnvironmentHandler(environments port.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{environments: environments}
}

func (h *EnvironmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateEnvironmentRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	e, err := h.environments.Create(r.Context(), req.Name, model.EnvironmentType(req.Type), req.SortOrder)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (h *EnvironmentHandler) List(w http.ResponseWriter, r *http.Request) {
	envs, err := h.environments.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envs)
}

func (h *EnvironmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "environmentId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.UpdateEnvironmentRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	e, err := h.environments.Update(r.Context(), id, req.Name, model.EnvironmentType(req.Type), req.Enabled)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *EnvironmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "environmentId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	if err := h.environments.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
