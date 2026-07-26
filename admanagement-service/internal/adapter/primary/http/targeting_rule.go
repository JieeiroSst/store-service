package http

import (
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

func (h *Handler) CreateTargetingRule(w http.ResponseWriter, r *http.Request) {
	var req model.AdTargetingRule
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.targetingRule.Create(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) GetTargetingRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	rule, err := h.targetingRule.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) UpdateTargetingRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.AdTargetingRule
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	req.ID = id

	if err := h.targetingRule.Update(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (h *Handler) DeleteTargetingRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.targetingRule.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ListTargetingRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.targetingRule.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}
