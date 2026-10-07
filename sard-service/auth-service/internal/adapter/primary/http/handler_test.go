package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/auth-service/config"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/clock"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/otp"
	"github.com/JIeeiroSst/auth-service/internal/application"
	"github.com/JIeeiroSst/auth-service/internal/domain"
)

type cards struct{}

func (cards) ResolvePAN(context.Context, string, string) (domain.CardRef, error) {
	return domain.CardRef{CardID: "card-1", UserID: 7, MaskedPAN: "730100******1234", Status: "NORMAL"}, nil
}

type sessions struct{}

func (sessions) Validate(_ context.Context, token string) (int64, error) {
	if token == "alice-session" {
		return 7, nil
	}
	return 0, domain.ErrUnauthorized
}

type users struct{}

func (users) Username(context.Context, int64) (string, error) { return "anguyen", nil }

type otpProvider struct{ n int }

func (p *otpProvider) Issue(context.Context, string) (string, time.Time, error) {
	p.n++
	return []string{"", "482913", "771204", "305518"}[p.n], time.Now().Add(time.Minute), nil
}

func (p *otpProvider) Verify(_ context.Context, _, code string) (bool, error) {
	return code == []string{"", "482913", "771204", "305518"}[p.n], nil
}

type healthy struct{}

func (healthy) Ping(context.Context) error { return nil }

func newAPI(t *testing.T, verifyPerMinute int) *httptest.Server {
	t.Helper()
	cfg := config.Defaults()
	cfg.Server.AccessTokens = []string{"merchant"}
	cfg.Server.InternalTokens = []string{"internal"}
	cfg.Server.VerifyPerMinute = verifyPerMinute
	svc := application.NewAuthenticationService(application.Settings{
		Policy:    domain.Policy{OTPTTL: 5 * time.Minute, MaxAttempts: 3, MaxResends: 2, ValidityTime: 10 * time.Minute},
		ExposeOTP: true,
	}, memory.NewChallenges(), memory.NewLocker(), cards{}, sessions{}, users{}, &otpProvider{}, otp.LogSender{},
		clock.System{}, metrics.NewPrometheus())
	srv := httptest.NewServer(NewRouter(NewHandler(cfg, svc, healthy{})))
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
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatal("security headers missing")
	}
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

func TestPaymentAuthenticationJourney(t *testing.T) {
	srv := newAPI(t, 100)
	init := map[string]any{"pan": "7301001234561234", "expiry": "10/31", "amount": 1500000, "currency": "VND", "merchant": "Shopee"}
	if s := call(t, srv, "POST", "/api/v1/payment-authentications", "", init, nil); s != 401 {
		t.Fatalf("merchant token required: %d", s)
	}
	var ch challengeDTO
	if s := call(t, srv, "POST", "/api/v1/payment-authentications", "merchant", init, &ch); s != 201 || ch.Status != "PENDING" || len(ch.OTP) != 6 || ch.AttemptsLeft != 3 {
		t.Fatalf("initiate: %d %+v", s, ch)
	}
	id, code := ch.ID, ch.OTP

	var mine struct{ Count int }
	if s := call(t, srv, "GET", "/api/v1/me/payment-authentications", "alice-session", nil, &mine); s != 200 || mine.Count != 1 {
		t.Fatalf("list mine: %d %+v", s, mine)
	}
	if s := call(t, srv, "GET", "/api/v1/me/payment-authentications", "", nil, nil); s != 401 {
		t.Fatalf("no session: %d", s)
	}
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}
	var mismatch struct {
		Error                 string
		PaymentAuthentication challengeDTO `json:"payment_authentication"`
	}
	if s := call(t, srv, "POST", "/api/v1/me/payment-authentications/"+id+"/verify", "alice-session", map[string]string{"otp": wrong}, &mismatch); s != 422 || mismatch.PaymentAuthentication.AttemptsLeft != 2 {
		t.Fatalf("wrong otp: %d %+v", s, mismatch)
	}
	var resent challengeDTO
	if s := call(t, srv, "POST", "/api/v1/me/payment-authentications/"+id+"/resend", "alice-session", nil, &resent); s != 200 || resent.Resends != 1 || resent.OTP == code {
		t.Fatalf("resend: %d %+v", s, resent)
	}
	code = resent.OTP
	if s := call(t, srv, "POST", "/api/v1/me/payment-authentications/"+id+"/verify", "alice-session", map[string]string{"otp": code}, &ch); s != 200 || ch.Status != "AUTHENTICATED" || ch.ValidUntil == nil {
		t.Fatalf("verify: %d %+v", s, ch)
	}
	consume := map[string]any{"card_id": "card-1", "amount": 1500000, "currency": "VND"}
	if s := call(t, srv, "POST", "/internal/v1/payment-authentications/"+id+"/consume", "merchant", consume, nil); s != 401 {
		t.Fatalf("internal token required: %d", s)
	}
	if s := call(t, srv, "POST", "/internal/v1/payment-authentications/"+id+"/consume", "internal", consume, &ch); s != 200 || ch.Status != "USED" {
		t.Fatalf("consume: %d %+v", s, ch)
	}
	if s := call(t, srv, "POST", "/internal/v1/payment-authentications/"+id+"/consume", "internal", consume, nil); s != 409 {
		t.Fatalf("replay: %d", s)
	}
	var fetched challengeDTO
	if s := call(t, srv, "GET", "/api/v1/payment-authentications/"+id, "merchant", nil, &fetched); s != 200 || fetched.Status != "USED" || fetched.OTP != "" {
		t.Fatalf("get: %d %+v", s, fetched)
	}
	if s := call(t, srv, "GET", "/api/v1/payment-authentications/nope", "merchant", nil, nil); s != 404 {
		t.Fatalf("missing: %d", s)
	}
}

func TestVerifyRateLimit(t *testing.T) {
	srv := newAPI(t, 2)
	for i, want := range []int{404, 404, 429} {
		if s := call(t, srv, "POST", "/api/v1/me/payment-authentications/x/verify", "alice-session", map[string]string{"otp": "123456"}, nil); s != want {
			t.Fatalf("call %d: %d want %d", i, s, want)
		}
	}
}

func TestLimiterWindow(t *testing.T) {
	l := newLimiter(1, time.Minute)
	now := time.Now()
	l.now = func() time.Time { return now }
	if !l.allow("k") || l.allow("k") {
		t.Fatal("limit of one per window")
	}
	now = now.Add(time.Minute)
	if !l.allow("k") {
		t.Fatal("window resets")
	}
}
