package http

import (
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

func (h *Handler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var req model.AdPosition
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.position.Create(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) GetPosition(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	position, err := h.position.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, position)
}

func (h *Handler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.AdPosition
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	req.ID = id

	if err := h.position.Update(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (h *Handler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.position.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ListPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.position.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, positions)
}
