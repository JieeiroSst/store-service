package http

import "net/http"

func (h *Handler) GetClick(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	click, err := h.click.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, click)
}

func (h *Handler) ListClicks(w http.ResponseWriter, r *http.Request) {
	clicks, err := h.click.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clicks)
}

func (h *Handler) ListClicksByAd(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	clicks, err := h.click.ListByAd(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clicks)
}
