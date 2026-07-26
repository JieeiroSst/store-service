package http

import (
	"net/http"
	"time"
)

func (h *Handler) GetPerformanceSummary(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	summary, err := h.performanceSummary.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) ListPerformanceSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.performanceSummary.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

// RecomputePerformance rolls the day's ad_impressions/ad_clicks counts into
// ad_performance_summary for the given ad and ?date=YYYY-MM-DD (default:
// today), computing CTR in the process.
func (h *Handler) RecomputePerformance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	date := parseDate(r.URL.Query().Get("date"), time.Now())

	summary, err := h.performanceSummary.Recompute(r.Context(), id, date)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
