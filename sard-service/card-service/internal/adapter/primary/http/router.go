package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/programs", h.Programs)

	api.HandleFunc("POST /api/v1/accounts", h.OpenAccount)
	api.HandleFunc("GET /api/v1/accounts/{id}", h.GetAccount)
	api.HandleFunc("GET /api/v1/customers/{customerId}/accounts", h.CustomerAccounts)
	api.HandleFunc("POST /api/v1/accounts/{id}/block", h.BlockAccount)
	api.HandleFunc("POST /api/v1/accounts/{id}/unblock", h.UnblockAccount)
	api.HandleFunc("POST /api/v1/accounts/{id}/cancel", h.CancelAccount)
	api.HandleFunc("PUT /api/v1/accounts/{id}/credit-limit", h.SetCreditLimit)
	api.HandleFunc("POST /api/v1/accounts/{id}/payments", h.Payment)
	api.HandleFunc("GET /api/v1/accounts/{id}/transactions", h.Transactions)
	api.HandleFunc("POST /api/v1/accounts/{id}/cards", h.IssueCard)
	api.HandleFunc("GET /api/v1/accounts/{id}/cards", h.AccountCards)

	api.HandleFunc("GET /api/v1/cards/{id}", h.GetCard)
	api.HandleFunc("POST /api/v1/cards/{id}/activate", h.ActivateCard)
	api.HandleFunc("POST /api/v1/cards/{id}/block", h.BlockCard)
	api.HandleFunc("POST /api/v1/cards/{id}/unblock", h.UnblockCard)
	api.HandleFunc("POST /api/v1/cards/{id}/cancel", h.CancelCard)
	api.HandleFunc("POST /api/v1/cards/{id}/report", h.ReportCard)
	api.HandleFunc("PUT /api/v1/cards/{id}/limits", h.UpdateLimits)
	api.HandleFunc("PUT /api/v1/cards/{id}/controls", h.UpdateControls)
	api.HandleFunc("PUT /api/v1/cards/{id}/pin", h.SetPIN)
	api.HandleFunc("GET /api/v1/cards/{id}/authorizations", h.CardAuthorizations)

	api.HandleFunc("POST /api/v1/authorizations", h.Authorize)
	api.HandleFunc("GET /api/v1/authorizations/{id}", h.GetAuthorization)
	api.HandleFunc("POST /api/v1/authorizations/{id}/confirm", h.ConfirmAuthorization)
	api.HandleFunc("POST /api/v1/authorizations/{id}/cancel", h.CancelAuthorization)

	internal := http.NewServeMux()
	internal.HandleFunc("POST /internal/v1/cards/resolve", h.ResolveCard)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("/api/", h.tokens.require(api))
	mux.Handle("/internal/", h.internal.require(internal))
	return mux
}
