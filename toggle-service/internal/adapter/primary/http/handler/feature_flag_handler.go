package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/dto"
	adminmw "github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type FeatureFlagHandler struct {
	flags port.FeatureFlagService
}

func NewFeatureFlagHandler(flags port.FeatureFlagService) *FeatureFlagHandler {
	return &FeatureFlagHandler{flags: flags}
}

func projectIDParam(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "projectId"))
}

func (h *FeatureFlagHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.CreateFlagRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	userID, _ := adminmw.UserID(r.Context())

	flag, err := h.flags.Create(r.Context(), port.CreateFlagInput{
		ProjectID:   projectID,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Type:        model.FeatureFlagType(req.Type),
	}, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, flag)
}

func (h *FeatureFlagHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	flags, err := h.flags.List(r.Context(), projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, flags)
}

func (h *FeatureFlagHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	flag, err := h.flags.Get(r.Context(), projectID, chi.URLParam(r, "key"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *FeatureFlagHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.UpdateFlagRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	userID, _ := adminmw.UserID(r.Context())

	flag, err := h.flags.Update(r.Context(), projectID, chi.URLParam(r, "key"), port.UpdateFlagInput{
		Name:        req.Name,
		Description: req.Description,
		Type:        model.FeatureFlagType(req.Type),
	}, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *FeatureFlagHandler) Archive(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	userID, _ := adminmw.UserID(r.Context())
	if err := h.flags.Archive(r.Context(), projectID, chi.URLParam(r, "key"), userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FeatureFlagHandler) toggle(w http.ResponseWriter, r *http.Request, enabled bool) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	userID, _ := adminmw.UserID(r.Context())
	err = h.flags.Toggle(r.Context(), projectID, chi.URLParam(r, "key"), chi.URLParam(r, "envName"), enabled, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FeatureFlagHandler) ToggleOn(w http.ResponseWriter, r *http.Request) {
	h.toggle(w, r, true)
}

func (h *FeatureFlagHandler) ToggleOff(w http.ResponseWriter, r *http.Request) {
	h.toggle(w, r, false)
}
