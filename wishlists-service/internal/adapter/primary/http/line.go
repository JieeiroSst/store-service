package http

import (
	"net/http"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
)

type updateLineRequest struct {
	Quantity int `json:"quantity"`
}

func (h *Handler) AddLine(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in port.AddLineInput
	if !decode(w, r, &in) {
		return
	}
	out, err := h.wishlists.AddLine(r.Context(), actor, id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, out)
}

func (h *Handler) UpdateLine(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	lineID, ok := pathID(w, r, "lineId")
	if !ok {
		return
	}
	var in updateLineRequest
	if !decode(w, r, &in) {
		return
	}
	out, err := h.wishlists.UpdateLine(r.Context(), actor, id, lineID, in.Quantity)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) RemoveLine(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	lineID, ok := pathID(w, r, "lineId")
	if !ok {
		return
	}
	if err := h.wishlists.RemoveLine(r.Context(), actor, id, lineID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
