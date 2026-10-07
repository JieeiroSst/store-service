package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-info-service/config"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/crypto"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/customer-info-service/internal/application"
	"github.com/JIeeiroSst/customer-info-service/internal/domain"
)

type clock struct{}

func (clock) Now() time.Time { return time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC) }

type users struct{}

func (users) GetUser(_ context.Context, id int64) (domain.UserProfile, error) {
	if id != 7 {
		return domain.UserProfile{}, domain.ErrUserNotFound
	}
	return domain.UserProfile{ID: 7, Username: "an", Name: "Nguyen Van An", Email: "an@mail.com", Active: true}, nil
}

type ekyc struct{}

func (e ekyc) SubmitCitizenCard(ctx context.Context, userID int64, _, _ []byte) (domain.EkycResult, error) {
	return e.Status(ctx, userID)
}

func (ekyc) Status(context.Context, int64) (domain.EkycResult, error) {
	d := time.Date(1995, 3, 4, 0, 0, 0, 0, time.UTC)
	return domain.EkycResult{
		Document:     &domain.IdentityDocument{Source: "card_ocr", Number: "079095001234", FullName: "NGUYEN VAN AN", DateOfBirth: &d, ChecksumValid: true, Confidence: 0.8},
		FaceVerified: true,
	}, nil
}

type healthy struct{}

func (healthy) Ping(context.Context) error { return nil }

func newAPI(t *testing.T, tokens ...string) *httptest.Server {
	t.Helper()
	svc := application.NewCustomerService(domain.KYCPolicy{MinAge: 18, MinConfidence: 0.5}, memory.NewCustomers(), memory.NewLocker(),
		users{}, ekyc{}, crypto.NewHasher([]byte("http-test-document-hash-key-0123456789")), clock{}, metrics.NewPrometheus())
	cfg := config.Defaults()
	cfg.Server.AccessTokens = tokens
	srv := httptest.NewServer(NewRouter(NewHandler(cfg, svc, healthy{}, clock{})))
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, req *http.Request, out any) int {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

func jsonReq(t *testing.T, method, url, body string) *http.Request {
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	return req
}

func imageReq(t *testing.T, url string, fields ...string) *http.Request {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for _, f := range fields {
		fw, _ := mw.CreateFormFile(f, f+".jpg")
		fw.Write([]byte("fake-image"))
	}
	mw.Close()
	req, _ := http.NewRequest("POST", url, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestCustomerJourney(t *testing.T) {
	srv := newAPI(t)
	var c customerDTO
	if s := do(t, jsonReq(t, "POST", srv.URL+"/api/v1/customers", `{"user_id":7}`), &c); s != 201 || c.FullName != "Nguyen Van An" || c.KYC.Status != "NONE" || c.CardEligibility.Eligible {
		t.Fatalf("onboard: %d %+v", s, c)
	}
	if s := do(t, jsonReq(t, "POST", srv.URL+"/api/v1/customers", `{"user_id":7}`), &c); s != 200 {
		t.Fatalf("second onboard: %d", s)
	}
	var e map[string]string
	if s := do(t, jsonReq(t, "POST", srv.URL+"/api/v1/customers", `{"user_id":9}`), &e); s != 404 || !strings.Contains(e["error"], "user-service") {
		t.Fatalf("unknown user: %d %v", s, e)
	}

	raw := map[string]any{}
	if s := do(t, imageReq(t, srv.URL+"/api/v1/customers/"+c.ID+"/kyc", "front", "back"), &raw); s != 200 {
		t.Fatalf("kyc: %d %v", s, raw)
	}
	kyc := raw["kyc"].(map[string]any)
	doc := kyc["document"].(map[string]any)
	if kyc["status"] != "VERIFIED" || kyc["face_verified"] != true || doc["number_last4"] != "1234" || doc["date_of_birth"] != "04/03/1995" {
		t.Fatalf("kyc body: %v", raw)
	}
	if strings.Contains(fmtJSON(raw), "079095001234") {
		t.Fatal("full document number leaked")
	}
	if raw["card_eligibility"].(map[string]any)["eligible"] != true {
		t.Fatalf("eligibility: %v", raw["card_eligibility"])
	}

	if s := do(t, jsonReq(t, "GET", srv.URL+"/api/v1/users/7/customer", ""), &c); s != 200 || c.UserID != 7 {
		t.Fatalf("by user: %d", s)
	}
	if s := do(t, jsonReq(t, "POST", srv.URL+"/api/v1/customers/"+c.ID+"/sync", ""), &c); s != 200 {
		t.Fatalf("sync: %d", s)
	}
	if s := do(t, jsonReq(t, "POST", srv.URL+"/api/v1/customers/"+c.ID+"/kyc/refresh", ""), &raw); s != 200 {
		t.Fatalf("refresh: %d %v", s, raw)
	}
	if s := do(t, imageReq(t, srv.URL+"/api/v1/customers/"+c.ID+"/kyc", "front"), &e); s != 400 {
		t.Fatalf("missing back: %d", s)
	}
	if s := do(t, jsonReq(t, "GET", srv.URL+"/api/v1/customers/nope", ""), &e); s != 404 {
		t.Fatalf("missing: %d", s)
	}
	if s := do(t, jsonReq(t, "POST", srv.URL+"/api/v1/customers/"+c.ID+"/kyc", "{}"), &e); s != 400 {
		t.Fatalf("no image: %d", s)
	}
}

func TestAccessTokens(t *testing.T) {
	srv := newAPI(t, "tok")
	if s := do(t, jsonReq(t, "GET", srv.URL+"/api/v1/customers/x", ""), nil); s != 401 {
		t.Fatalf("no token: %d", s)
	}
	req := jsonReq(t, "GET", srv.URL+"/api/v1/customers/x", "")
	req.Header.Set("Authorization", "Bearer tok")
	if s := do(t, req, nil); s != 404 {
		t.Fatalf("with token: %d", s)
	}
	if s := do(t, jsonReq(t, "GET", srv.URL+"/health", ""), nil); s != 200 {
		t.Fatalf("health: %d", s)
	}
}

func fmtJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
