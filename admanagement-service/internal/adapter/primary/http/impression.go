package http

import "net/http"

func (h *Handler) GetImpression(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	impression, err := h.impression.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, impression)
}

func (h *Handler) ListImpressions(w http.ResponseWriter, r *http.Request) {
	impressions, err := h.impression.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, impressions)
}

func (h *Handler) ListImpressionsByAd(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	impressions, err := h.impression.ListByAd(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, impressions)
}
