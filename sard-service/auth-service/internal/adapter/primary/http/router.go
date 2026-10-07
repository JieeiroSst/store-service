package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)

	mux.HandleFunc("POST /api/v1/payment-authentications", h.merchant.require(h.Initiate))
	mux.HandleFunc("GET /api/v1/payment-authentications/{id}", h.merchant.require(h.Get))

	mux.HandleFunc("GET /api/v1/me/payment-authentications", h.ListMine)
	mux.HandleFunc("POST /api/v1/me/payment-authentications/{id}/verify", h.limiter.wrap(h.Verify))
	mux.HandleFunc("POST /api/v1/me/payment-authentications/{id}/resend", h.limiter.wrap(h.Resend))
	mux.HandleFunc("POST /api/v1/me/payment-authentications/{id}/decline", h.limiter.wrap(h.Decline))

	mux.HandleFunc("POST /internal/v1/payment-authentications/{id}/consume", h.internal.require(h.Consume))
	return securityHeaders(mux)
}
