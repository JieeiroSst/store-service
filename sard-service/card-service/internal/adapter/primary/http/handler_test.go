package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/card-service/config"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/authservice"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/clock"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/crypto"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/card-service/internal/application"
	"github.com/JIeeiroSst/card-service/internal/domain"
)

type healthy struct{}

func (healthy) Ping(context.Context) error { return nil }

type directory struct{}

func (directory) GetCustomer(_ context.Context, id string) (domain.Customer, error) {
	return domain.Customer{ID: id, UserID: 42, FullName: "Lê Văn Bình", Eligible: true}, nil
}

func newAPI(t *testing.T) *httptest.Server {
	t.Helper()
	keys := crypto.DeriveKeys([]byte("http-test-master-key-0123456789abcdef"))
	vault, _ := crypto.NewVault(keys)
	sec := crypto.NewSecurity(keys)
	store := memory.NewStore()
	accounts, cards, auths, txns := memory.NewAccounts(store), memory.NewCards(store), memory.NewAuthorizations(store), memory.NewTransactions(store)
	locker := memory.NewLocker()
	m := metrics.NewPrometheus()
	clk := clock.System{}
	catalog := domain.NewCatalog(domain.Programs)
	accountSvc := application.NewAccountService(catalog, accounts, cards, txns, directory{}, locker, clk, m)
	cardSvc := application.NewCardService(catalog, accounts, cards, locker, vault, sec, clk, m)
	authSvc := application.NewAuthorizationService(application.Settings{LimitsLocation: time.UTC}, catalog, accounts, cards, auths, txns, authservice.Disabled{}, locker, vault, sec, clk, m)
	cfg := config.Defaults()
	cfg.Server.AccessTokens = []string{"api"}
	cfg.Server.InternalTokens = []string{"internal"}
	srv := httptest.NewServer(NewRouter(NewHandler(cfg, accountSvc, cardSvc, authSvc, healthy{})))
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path, token string, body any, out any) int {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, srv.URL+path, &buf)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("%s %s: decode: %v", method, path, err)
		}
	}
	return resp.StatusCode
}

func TestAccountCardAuthorizationJourney(t *testing.T) {
	srv := newAPI(t)
	var programs struct{ Count int }
	if s := call(t, srv, "GET", "/api/v1/programs", "api", nil, &programs); s != 200 || programs.Count != len(domain.Programs) {
		t.Fatalf("programs: %d %+v", s, programs)
	}
	if s := call(t, srv, "GET", "/api/v1/programs", "", nil, nil); s != 401 {
		t.Fatalf("token required: %d", s)
	}

	var acc accountDTO
	if s := call(t, srv, "POST", "/api/v1/accounts", "api", map[string]any{"customer_id": "cus-100", "program_code": "debit_classic"}, &acc); s != 201 || acc.Status != "NORMAL" || acc.Mode != "DEBIT" {
		t.Fatalf("open: %d %+v", s, acc)
	}
	var paid struct {
		Account     accountDTO
		Transaction transactionDTO
	}
	if s := call(t, srv, "POST", "/api/v1/accounts/"+acc.ID+"/payments", "api", map[string]any{"amount": 5_000_000, "description": "top up"}, &paid); s != 201 || paid.Account.Available != 5_000_000 || paid.Transaction.ProcessingCode != "28" {
		t.Fatalf("payment: %d %+v", s, paid)
	}

	var issued issuedCardDTO
	if s := call(t, srv, "POST", "/api/v1/accounts/"+acc.ID+"/cards", "api", map[string]any{"type": "plastic"}, &issued); s != 201 || issued.Card.Status != "PENDING" || issued.Card.CardholderName != "LE VAN BINH" {
		t.Fatalf("issue: %d %+v", s, issued)
	}
	id := issued.Card.ID
	var errBody map[string]string
	if s := call(t, srv, "POST", "/api/v1/cards/"+id+"/activate", "api", map[string]string{"cvv": "abc"}, &errBody); s != 422 {
		t.Fatalf("bad cvv: %d %v", s, errBody)
	}
	var card cardDTO
	if s := call(t, srv, "POST", "/api/v1/cards/"+id+"/activate", "api", map[string]string{"cvv": issued.CVV}, &card); s != 200 || card.Status != "NORMAL" {
		t.Fatalf("activate: %d %+v", s, card)
	}

	var resolved resolvedCardDTO
	resolveReq := map[string]string{"pan": issued.PAN, "expiry": issued.Card.Expiry}
	if s := call(t, srv, "POST", "/internal/v1/cards/resolve", "api", resolveReq, nil); s != 401 {
		t.Fatalf("resolve needs the internal token: %d", s)
	}
	if s := call(t, srv, "POST", "/internal/v1/cards/resolve", "internal", resolveReq, &resolved); s != 200 || resolved.UserID != 42 || resolved.Status != "NORMAL" || resolved.CardID != id {
		t.Fatalf("resolve: %d %+v", s, resolved)
	}

	var auth authorizationDTO
	authReq := map[string]any{
		"pan": issued.PAN, "expiry": issued.Card.Expiry, "cvv": issued.CVV, "amount": 3_000_000,
		"currency": "VND", "channel": "ECOM", "merchant": "Shopee", "mcc": "5311",
	}
	if s := call(t, srv, "POST", "/api/v1/authorizations", "api", authReq, &auth); s != 200 || !auth.Approved || auth.Status != "AUTHORIZED" || auth.ProcessingCode != "00" {
		t.Fatalf("authorize: %d %+v", s, auth)
	}
	held := auth.ID
	var declined authorizationDTO
	if s := call(t, srv, "POST", "/api/v1/authorizations", "api", authReq, &declined); s != 200 || declined.Approved || declined.ResponseCode != "51" {
		t.Fatalf("insufficient funds: %d %+v", s, declined)
	}
	var confirmed authorizationDTO
	if s := call(t, srv, "POST", "/api/v1/authorizations/"+held+"/confirm", "api", map[string]int64{"amount": 2_500_000}, &confirmed); s != 200 || confirmed.Status != "CONFIRMED" || confirmed.ConfirmedAmount != 2_500_000 {
		t.Fatalf("confirm: %d %+v", s, confirmed)
	}
	if s := call(t, srv, "POST", "/api/v1/authorizations/"+held+"/cancel", "api", nil, &errBody); s != 409 {
		t.Fatalf("cancel confirmed: %d", s)
	}
	if s := call(t, srv, "GET", "/api/v1/accounts/"+acc.ID, "api", nil, &acc); s != 200 || acc.Balance != 2_500_000 || acc.Held != 0 {
		t.Fatalf("account after clearing: %d %+v", s, acc)
	}
	var txns struct {
		Transactions []transactionDTO
		Count        int
	}
	if s := call(t, srv, "GET", "/api/v1/accounts/"+acc.ID+"/transactions", "api", nil, &txns); s != 200 || txns.Count != 2 || txns.Transactions[0].Amount != -2_500_000 {
		t.Fatalf("transactions: %d %+v", s, txns)
	}
	if s := call(t, srv, "GET", "/api/v1/cards/"+id, "api", nil, &card); s != 200 || card.SpentToday == nil || *card.SpentToday != 2_500_000 {
		t.Fatalf("spent today: %d %+v", s, card)
	}

	var reported struct {
		Card        cardDTO
		Replacement *issuedCardDTO
	}
	if s := call(t, srv, "POST", "/api/v1/cards/"+id+"/report", "api", map[string]any{"reason": "robbed", "reissue": true}, &reported); s != 200 || reported.Card.Status != "ROBBED" || reported.Replacement == nil {
		t.Fatalf("report: %d %+v", s, reported)
	}
	var cards struct{ Count int }
	if s := call(t, srv, "GET", "/api/v1/accounts/"+acc.ID+"/cards", "api", nil, &cards); s != 200 || cards.Count != 2 {
		t.Fatalf("account cards: %d %+v", s, cards)
	}
	var accounts struct{ Count int }
	if s := call(t, srv, "GET", "/api/v1/customers/cus-100/accounts", "api", nil, &accounts); s != 200 || accounts.Count != 1 {
		t.Fatalf("customer accounts: %d %+v", s, accounts)
	}
}

func TestErrors(t *testing.T) {
	srv := newAPI(t)
	var body map[string]string
	cases := []struct {
		method, path string
		body         any
		want         int
	}{
		{"GET", "/api/v1/cards/nope", nil, 404},
		{"GET", "/api/v1/accounts/nope", nil, 404},
		{"POST", "/api/v1/accounts", map[string]string{"customer_id": "c", "program_code": "X"}, 400},
		{"POST", "/api/v1/accounts", map[string]string{"unknown": "field"}, 400},
		{"POST", "/api/v1/accounts/nope/cards", map[string]string{"type": "METAL"}, 400},
		{"POST", "/api/v1/authorizations", map[string]any{"pan": "123", "expiry": "01/30", "amount": 1, "currency": "VND", "channel": "POS"}, 400},
		{"POST", "/api/v1/authorizations/nope/confirm", nil, 404},
		{"GET", "/api/v1/accounts/x/transactions?limit=-1", nil, 400},
	}
	for _, c := range cases {
		if got := call(t, srv, c.method, c.path, "api", c.body, &body); got != c.want {
			t.Errorf("%s %s = %d (%v), want %d", c.method, c.path, got, body, c.want)
		}
	}
	if s := call(t, srv, "GET", "/health", "", nil, nil); s != 200 {
		t.Fatalf("health: %d", s)
	}
}
