package http

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/card-service/config"
	"github.com/JIeeiroSst/card-service/internal/application"
	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/JIeeiroSst/card-service/internal/port"
)

const (
	maxBodyBytes  = 64 << 10
	healthTimeout = 2 * time.Second
)

type Handler struct {
	accounts *application.AccountService
	cards    *application.CardService
	auths    *application.AuthorizationService
	health   port.HealthChecker
	tokens   tokenSet
	internal tokenSet
}

func NewHandler(cfg *config.Config, accounts *application.AccountService, cards *application.CardService, auths *application.AuthorizationService, health port.HealthChecker) *Handler {
	return &Handler{
		accounts: accounts,
		cards:    cards,
		auths:    auths,
		health:   health,
		tokens:   newTokenSet(cfg.Server.AccessTokens),
		internal: newTokenSet(cfg.Server.InternalTokens),
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()
	if err := h.health.Ping(ctx); err != nil {
		log.Printf("health check failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "version": config.Version})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": config.Version})
}

func (h *Handler) Programs(w http.ResponseWriter, _ *http.Request) {
	programs := h.accounts.Programs()
	out := make([]programDTO, 0, len(programs))
	for _, p := range programs {
		out = append(out, toProgram(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"programs": out, "count": len(out)})
}

func (h *Handler) OpenAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID  string `json:"customer_id"`
		ProgramCode string `json:"program_code"`
		CreditLimit int64  `json:"credit_limit"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	a, err := h.accounts.Open(r.Context(), application.OpenAccountCommand{
		CustomerID:  strings.TrimSpace(req.CustomerID),
		ProgramCode: strings.ToUpper(strings.TrimSpace(req.ProgramCode)),
		CreditLimit: req.CreditLimit,
	})
	respond(w, http.StatusCreated, a, err, toAccount)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	a, err := h.accounts.Get(r.Context(), r.PathValue("id"))
	respond(w, http.StatusOK, a, err, toAccount)
}

func (h *Handler) CustomerAccounts(w http.ResponseWriter, r *http.Request) {
	list, err := h.accounts.ListByCustomer(r.Context(), r.PathValue("customerId"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": mapList(list, toAccount), "count": len(list)})
}

func (h *Handler) BlockAccount(w http.ResponseWriter, r *http.Request) {
	var req reasonRequest
	if err := decodeOptionalBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	a, err := h.accounts.Block(r.Context(), r.PathValue("id"), req.Reason)
	respond(w, http.StatusOK, a, err, toAccount)
}

func (h *Handler) UnblockAccount(w http.ResponseWriter, r *http.Request) {
	a, err := h.accounts.Unblock(r.Context(), r.PathValue("id"))
	respond(w, http.StatusOK, a, err, toAccount)
}

func (h *Handler) CancelAccount(w http.ResponseWriter, r *http.Request) {
	var req reasonRequest
	if err := decodeOptionalBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	a, err := h.accounts.Cancel(r.Context(), r.PathValue("id"), req.Reason)
	respond(w, http.StatusOK, a, err, toAccount)
}

func (h *Handler) SetCreditLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CreditLimit int64 `json:"credit_limit"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	a, err := h.accounts.SetCreditLimit(r.Context(), r.PathValue("id"), req.CreditLimit)
	respond(w, http.StatusOK, a, err, toAccount)
}

func (h *Handler) Payment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount      int64  `json:"amount"`
		Description string `json:"description"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	a, txn, err := h.accounts.ReceivePayment(r.Context(), r.PathValue("id"), req.Amount, req.Description)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"account": toAccount(a), "transaction": toTransaction(txn)})
}

func (h *Handler) Transactions(w http.ResponseWriter, r *http.Request) {
	limit, err := limitParam(r)
	if err != nil {
		fail(w, err)
		return
	}
	list, err := h.accounts.Transactions(r.Context(), r.PathValue("id"), limit)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": mapList(list, toTransaction), "count": len(list)})
}

func (h *Handler) IssueCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type        string `json:"type"`
		PrintedName string `json:"printed_name"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	t, err := domain.ParseCardType(strings.ToUpper(strings.TrimSpace(req.Type)))
	if err != nil {
		fail(w, err)
		return
	}
	issued, err := h.cards.Issue(r.Context(), application.IssueCommand{AccountID: r.PathValue("id"), Type: t, PrintedName: req.PrintedName})
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, toIssued(issued))
}

func (h *Handler) AccountCards(w http.ResponseWriter, r *http.Request) {
	list, err := h.cards.ListByAccount(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cards": mapList(list, toCard), "count": len(list)})
}

func (h *Handler) GetCard(w http.ResponseWriter, r *http.Request) {
	card, err := h.cards.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	spent, err := h.auths.SpentToday(r.Context(), card.ID)
	if err != nil {
		fail(w, err)
		return
	}
	dto := toCard(card)
	dto.SpentToday = &spent
	writeJSON(w, http.StatusOK, dto)
}

func (h *Handler) ActivateCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CVV string `json:"cvv"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.cards.Activate(r.Context(), r.PathValue("id"), req.CVV)
	respond(w, http.StatusOK, c, err, toCard)
}

func (h *Handler) BlockCard(w http.ResponseWriter, r *http.Request) {
	var req reasonRequest
	if err := decodeOptionalBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.cards.Block(r.Context(), r.PathValue("id"), req.Reason)
	respond(w, http.StatusOK, c, err, toCard)
}

func (h *Handler) UnblockCard(w http.ResponseWriter, r *http.Request) {
	c, err := h.cards.Unblock(r.Context(), r.PathValue("id"))
	respond(w, http.StatusOK, c, err, toCard)
}

func (h *Handler) CancelCard(w http.ResponseWriter, r *http.Request) {
	var req reasonRequest
	if err := decodeOptionalBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.cards.Cancel(r.Context(), r.PathValue("id"), req.Reason)
	respond(w, http.StatusOK, c, err, toCard)
}

func (h *Handler) ReportCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason  string `json:"reason"`
		Note    string `json:"note"`
		Reissue bool   `json:"reissue"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	reason, err := domain.ParseReportReason(strings.ToUpper(strings.TrimSpace(req.Reason)))
	if err != nil {
		fail(w, err)
		return
	}
	card, replacement, err := h.cards.Report(r.Context(), r.PathValue("id"), application.ReportCommand{Reason: reason, Note: req.Note, Reissue: req.Reissue})
	if err != nil {
		fail(w, err)
		return
	}
	resp := map[string]any{"card": toCard(card)}
	if replacement != nil {
		w.Header().Set("Cache-Control", "no-store")
		resp["replacement"] = toIssued(replacement)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateLimits(w http.ResponseWriter, r *http.Request) {
	var req limitsDTO
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.cards.UpdateLimits(r.Context(), r.PathValue("id"), domain.Limits{PerTransaction: req.PerTransaction, Daily: req.Daily})
	respond(w, http.StatusOK, c, err, toCard)
}

func (h *Handler) UpdateControls(w http.ResponseWriter, r *http.Request) {
	var req controlsDTO
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.cards.UpdateControls(r.Context(), r.PathValue("id"), req.domain())
	respond(w, http.StatusOK, c, err, toCard)
}

func (h *Handler) SetPIN(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PIN string `json:"pin"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.cards.SetPIN(r.Context(), r.PathValue("id"), req.PIN)
	respond(w, http.StatusOK, c, err, toCard)
}

func (h *Handler) CardAuthorizations(w http.ResponseWriter, r *http.Request) {
	limit, err := limitParam(r)
	if err != nil {
		fail(w, err)
		return
	}
	list, err := h.auths.ListByCard(r.Context(), r.PathValue("id"), limit)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authorizations": mapList(list, toAuthorization), "count": len(list)})
}

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PAN              string `json:"pan"`
		Expiry           string `json:"expiry"`
		CVV              string `json:"cvv"`
		PIN              string `json:"pin"`
		Amount           int64  `json:"amount"`
		Currency         string `json:"currency"`
		Channel          string `json:"channel"`
		ProcessingCode   string `json:"processing_code"`
		Merchant         string `json:"merchant"`
		MCC              string `json:"mcc"`
		AuthenticationID string `json:"authentication_id"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	expiry, err := domain.ParseExpiry(req.Expiry)
	if err != nil {
		fail(w, err)
		return
	}
	a, err := h.auths.Authorize(r.Context(), domain.AuthorizationRequest{
		PAN: req.PAN, Expiry: expiry, CVV: req.CVV, PIN: req.PIN, Amount: req.Amount, Currency: req.Currency,
		Channel: domain.Channel(req.Channel), ProcessingCode: domain.ProcessingCode(req.ProcessingCode),
		Merchant: req.Merchant, MCC: req.MCC, AuthenticationID: req.AuthenticationID,
	})
	respond(w, http.StatusOK, a, err, toAuthorization)
}

func (h *Handler) GetAuthorization(w http.ResponseWriter, r *http.Request) {
	a, err := h.auths.Get(r.Context(), r.PathValue("id"))
	respond(w, http.StatusOK, a, err, toAuthorization)
}

func (h *Handler) ConfirmAuthorization(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := decodeOptionalBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	a, err := h.auths.Confirm(r.Context(), r.PathValue("id"), req.Amount)
	respond(w, http.StatusOK, a, err, toAuthorization)
}

func (h *Handler) CancelAuthorization(w http.ResponseWriter, r *http.Request) {
	a, err := h.auths.Cancel(r.Context(), r.PathValue("id"))
	respond(w, http.StatusOK, a, err, toAuthorization)
}

func (h *Handler) ResolveCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PAN    string `json:"pan"`
		Expiry string `json:"expiry"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	expiry, err := domain.ParseExpiry(req.Expiry)
	if err != nil {
		fail(w, err)
		return
	}
	res, err := h.cards.Resolve(r.Context(), req.PAN, expiry)
	if err != nil {
		fail(w, err)
		return
	}
	status := string(res.Card.Status)
	if res.Account.Status != domain.AccountNormal {
		status = string(domain.StatusBlocked)
	}
	writeJSON(w, http.StatusOK, resolvedCardDTO{
		CardID: res.Card.ID, AccountID: res.Account.ID, CustomerID: res.Account.CustomerID,
		UserID: res.Account.UserID, MaskedPAN: res.Card.MaskedPAN(), Status: status,
	})
}

type reasonRequest struct {
	Reason string `json:"reason"`
}

func respond[T, D any](w http.ResponseWriter, status int, v *T, err error, f func(*T) D) {
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, status, f(v))
}

func limitParam(r *http.Request) (int, error) {
	v := r.URL.Query().Get("limit")
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, domain.Invalid("limit must be a positive integer")
	}
	return n, nil
}

type tokenSet [][]byte

func newTokenSet(tokens []string) tokenSet {
	var out tokenSet
	for _, t := range tokens {
		out = append(out, []byte(t))
	}
	return out
}

func (ts tokenSet) require(next http.Handler) http.Handler {
	if len(ts) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ts.valid(token(r)) {
			fail(w, domain.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (ts tokenSet) valid(t string) bool {
	if t == "" {
		return false
	}
	ok := 0
	for _, want := range ts {
		ok |= subtle.ConstantTimeCompare([]byte(t), want)
	}
	return ok == 1
}

func token(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return r.Header.Get("X-Api-Token")
}

func fail(w http.ResponseWriter, err error) {
	status, msg := errorStatus(err)
	if status >= 500 {
		log.Printf("request failed: %v", err)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func errorStatus(err error) (int, string) {
	switch {
	case domain.IsInvalid(err):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "missing or invalid access token"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case domain.IsConflict(err):
		return http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrCVVMismatch):
		return http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, domain.ErrUnavailable):
		return http.StatusServiceUnavailable, "a dependency is unavailable, try again later"
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "request timed out"
	case errors.Is(err, context.Canceled):
		return http.StatusServiceUnavailable, "request canceled"
	}
	return http.StatusInternalServerError, "internal error"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeBody(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return domain.Invalid("invalid JSON body: %v", err)
	}
	return nil
}

func decodeOptionalBody(r *http.Request, v any) error {
	err := decodeBody(r, v)
	if err != nil && r.ContentLength == 0 {
		return nil
	}
	return err
}
