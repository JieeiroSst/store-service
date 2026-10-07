package secondary_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/cardservice"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/otp"
	"github.com/JIeeiroSst/auth-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/JIeeiroSst/auth-service/internal/port"
)

func TestUserServiceValidate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch body["session_token"] {
		case "good":
			w.Write([]byte(`{"valid":true,"userId":"7"}`))
		case "snake":
			w.Write([]byte(`{"valid":true,"user_id":"9"}`))
		case "down":
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.Write([]byte(`{"valid":false,"userId":""}`))
		}
	}))
	defer srv.Close()
	c := userservice.NewClient(srv.URL, "", time.Second)
	ctx := context.Background()
	if id, err := c.Validate(ctx, "good"); err != nil || id != 7 {
		t.Fatalf("good: %d %v", id, err)
	}
	if id, err := c.Validate(ctx, "snake"); err != nil || id != 9 {
		t.Fatalf("snake: %d %v", id, err)
	}
	if _, err := c.Validate(ctx, "bad"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("bad: %v", err)
	}
	if _, err := c.Validate(ctx, ""); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("empty: %v", err)
	}
	if _, err := c.Validate(ctx, "down"); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("down: %v", err)
	}
}

func TestCardServiceResolve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer internal" || r.URL.Path != "/internal/v1/cards/resolve" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["pan"] != "7301001234561234" || body["expiry"] != "10/31" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(`{"card_id":"c1","account_id":"a1","customer_id":"cu1","user_id":7,"masked_pan":"730100******1234","status":"NORMAL"}`))
	}))
	defer srv.Close()
	c := cardservice.NewClient(srv.URL, "internal", time.Second)
	ref, err := c.ResolvePAN(context.Background(), "7301001234561234", "10/31")
	if err != nil || ref.CardID != "c1" || ref.UserID != 7 || ref.Status != "NORMAL" {
		t.Fatalf("%v %+v", err, ref)
	}
	if _, err := c.ResolvePAN(context.Background(), "7301009999999999", "10/31"); !domain.IsConflict(err) {
		t.Fatalf("unknown: %v", err)
	}
	if _, err := cardservice.NewClient(srv.URL, "x", time.Second).ResolvePAN(context.Background(), "7301001234561234", "10/31"); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("bad token: %v", err)
	}
}

func TestUserServiceUsername(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer internal" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/internal/v1/users/7":
			w.Write([]byte(`{"id":7,"username":"anguyen","active":true}`))
		case "/internal/v1/users/8":
			w.Write([]byte(`{"id":8,"username":"locked","active":false}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := userservice.NewClient(srv.URL, "internal", time.Second)
	ctx := context.Background()
	if name, err := c.Username(ctx, 7); err != nil || name != "anguyen" {
		t.Fatalf("%q %v", name, err)
	}
	if _, err := c.Username(ctx, 8); !domain.IsConflict(err) {
		t.Fatalf("locked user: %v", err)
	}
	if _, err := c.Username(ctx, 9); !domain.IsConflict(err) {
		t.Fatalf("missing user: %v", err)
	}
	if _, err := userservice.NewClient(srv.URL, "", time.Second).Username(ctx, 7); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("no token: %v", err)
	}
}

func TestOTPWebhook(t *testing.T) {
	got := make(chan map[string]any, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.Header.Get("Authorization") == "Bearer hook" {
			got <- body
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	err := otp.NewWebhook(srv.URL, "hook", time.Second).Send(context.Background(), port.OTPMessage{
		UserID: 7, ChallengeID: "ch-1", OTP: "123456", Amount: 100, Currency: "VND", Merchant: "Tiki", ExpiresAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	body := <-got
	if body["user_id"].(float64) != 7 || body["otp"] != "123456" || body["template"] != "payment_authentication" {
		t.Fatalf("payload %v", body)
	}
}
