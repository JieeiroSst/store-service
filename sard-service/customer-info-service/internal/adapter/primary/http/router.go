package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("POST /api/v1/customers", h.Onboard)
	api.HandleFunc("GET /api/v1/customers/{id}", h.GetCustomer)
	api.HandleFunc("POST /api/v1/customers/{id}/sync", h.Sync)
	api.HandleFunc("POST /api/v1/customers/{id}/kyc", h.SubmitKYC)
	api.HandleFunc("POST /api/v1/customers/{id}/kyc/refresh", h.RefreshKYC)
	api.HandleFunc("GET /api/v1/users/{userId}/customer", h.GetByUser)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("/api/", h.authorize(api))
	return mux
}
