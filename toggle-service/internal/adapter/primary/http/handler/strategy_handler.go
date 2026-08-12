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

type StrategyHandler struct {
	strategies port.StrategyService
}

func NewStrategyHandler(strategies port.StrategyService) *StrategyHandler {
	return &StrategyHandler{strategies: strategies}
}

func toStrategyInput(req dto.StrategyRequest) port.StrategyInput {
	constraints := make([]port.ConstraintInput, 0, len(req.Constraints))
	for _, c := range req.Constraints {
		constraints = append(constraints, port.ConstraintInput{
			ContextField:    c.ContextField,
			Operator:        model.ConstraintOperator(c.Operator),
			Values:          c.Values,
			CaseInsensitive: c.CaseInsensitive,
		})
	}
	return port.StrategyInput{
		StrategyType: model.StrategyType(req.StrategyType),
		Parameters:   req.Parameters,
		SortOrder:    req.SortOrder,
		Constraints:  constraints,
	}
}

func (h *StrategyHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	strategies, err := h.strategies.List(r.Context(), projectID, chi.URLParam(r, "key"), chi.URLParam(r, "envName"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, strategies)
}

func (h *StrategyHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.StrategyRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	userID, _ := adminmw.UserID(r.Context())

	st, err := h.strategies.Add(r.Context(), projectID, chi.URLParam(r, "key"), chi.URLParam(r, "envName"), toStrategyInput(req), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, st)
}

func (h *StrategyHandler) Update(w http.ResponseWriter, r *http.Request) {
	strategyID, err := uuid.Parse(chi.URLParam(r, "strategyId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.StrategyRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	userID, _ := adminmw.UserID(r.Context())

	st, err := h.strategies.Update(r.Context(), strategyID, toStrategyInput(req), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (h *StrategyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	strategyID, err := uuid.Parse(chi.URLParam(r, "strategyId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	userID, _ := adminmw.UserID(r.Context())
	if err := h.strategies.Delete(r.Context(), strategyID, userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
