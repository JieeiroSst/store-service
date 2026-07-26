package http

import (
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

func (h *Handler) CreatePositionMapping(w http.ResponseWriter, r *http.Request) {
	var req model.AdPositionMapping
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.positionMapping.Create(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) GetPositionMapping(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	mapping, err := h.positionMapping.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapping)
}

func (h *Handler) UpdatePositionMapping(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.AdPositionMapping
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	req.ID = id

	if err := h.positionMapping.Update(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (h *Handler) DeletePositionMapping(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.positionMapping.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ListPositionMappings(w http.ResponseWriter, r *http.Request) {
	mappings, err := h.positionMapping.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mappings)
}
