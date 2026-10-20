package http

import "net/http"

func (h *Handler) CheckinStatus(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	st, err := h.checkins.Status(r.Context(), actor)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, st)
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	res, err := h.checkins.CheckIn(r.Context(), actor, clientInfo(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, res)
}

func (h *Handler) CheckinHistory(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	limit, offset, ok := paging(w, r)
	if !ok {
		return
	}
	p, err := h.checkins.History(r.Context(), actor, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, p)
}
