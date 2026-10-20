package http

import (
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) Routes() http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/lucky/games", h.Lobby)
	api.HandleFunc("GET /api/v1/lucky/games/{game}", h.CurrentCampaign)
	api.HandleFunc("GET /api/v1/lucky/games/{game}/winners", h.Winners)
	api.HandleFunc("GET /api/v1/lucky/games/{game}/me/status", h.MyStatus)
	api.HandleFunc("POST /api/v1/lucky/games/{game}/me/plays", h.Play)
	api.HandleFunc("GET /api/v1/lucky/me/plays", h.MyPlays)
	api.HandleFunc("GET /api/v1/lucky/checkin/me", h.CheckinStatus)
	api.HandleFunc("POST /api/v1/lucky/checkin/me", h.CheckIn)
	api.HandleFunc("GET /api/v1/lucky/checkin/me/history", h.CheckinHistory)
	api.HandleFunc("POST /api/v1/lucky/credits", h.GrantPlays)
	api.HandleFunc("POST /api/v1/lucky/campaigns", h.CreateCampaign)
	api.HandleFunc("GET /api/v1/lucky/campaigns", h.ListCampaigns)
	api.HandleFunc("GET /api/v1/lucky/campaigns/{id}", h.GetCampaign)
	api.HandleFunc("PUT /api/v1/lucky/campaigns/{id}", h.UpdateCampaign)
	api.HandleFunc("DELETE /api/v1/lucky/campaigns/{id}", h.DeleteCampaign)
	api.HandleFunc("POST /api/v1/lucky/campaigns/{id}/activate", h.ActivateCampaign)
	api.HandleFunc("POST /api/v1/lucky/campaigns/{id}/pause", h.PauseCampaign)
	api.HandleFunc("GET /api/v1/lucky/campaigns/{id}/plays", h.CampaignPlays)
	api.HandleFunc("POST /api/v1/lucky/fraud/blocks", h.BlockUser)
	api.HandleFunc("GET /api/v1/lucky/fraud/blocks", h.ListBlocks)
	api.HandleFunc("DELETE /api/v1/lucky/fraud/blocks/{userId}", h.UnblockUser)
	api.HandleFunc("GET /api/v1/lucky/fraud/rewards", h.ListRewards)
	api.HandleFunc("POST /api/v1/lucky/fraud/rewards/{id}/approve", h.ApproveReward)
	api.HandleFunc("POST /api/v1/lucky/fraud/rewards/{id}/reject", h.RejectReward)
	api.HandleFunc("GET /api/v1/lucky/fraud/plays", h.InvestigatePlays)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", h.Livez)
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("/api/", h.auth.OptionalMiddleware(api))
	return recoverer(accessLog(mux))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if r.URL.Path == "/livez" || r.URL.Path == "/health" {
			return
		}
		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "duration", time.Since(start))
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				slog.Error("panic while serving request", "panic", v, "path", r.URL.Path)
				writeJSON(w, http.StatusInternalServerError, errorBody("internal", "internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
