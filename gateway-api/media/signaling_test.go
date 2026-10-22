package media

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/gateway-api/auth"
)

func TestSignaling(t *testing.T) {
	var gotQuery, gotAuth, gotOrigin string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery, gotAuth, gotOrigin = r.URL.RawQuery, r.Header.Get("Authorization"), r.Header.Get("Origin")
		w.WriteHeader(http.StatusTeapot)
	}))
	defer upstream.Close()
	h := Signaling("webrtc-service", func(string) string { return upstream.URL })

	req := httptest.NewRequest(http.MethodGet, "/rtc/ws?room_id=team&user_id=someone-else", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/rtc/ws?room_id=team&user_id=someone-else", nil)
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{ID: "42", Token: "tok"}))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("authenticated: got %d", rec.Code)
	}
	if gotQuery != "room_id=team&user_id=42" || gotAuth != "" || gotOrigin != "" {
		t.Fatalf("query=%q auth=%q origin=%q", gotQuery, gotAuth, gotOrigin)
	}

	req = httptest.NewRequest(http.MethodGet, "/rtc/ws", nil)
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{ID: "42"}))
	req.Header.Set("Upgrade", "websocket")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing room: got %d", rec.Code)
	}
}

func TestSocket(t *testing.T) {
	var gotPath, gotQuery, gotAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery, gotAuth = r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization")
		w.WriteHeader(http.StatusTeapot)
	}))
	defer upstream.Close()
	h := Socket("ollama-service", "/ws", func(string) string { return upstream.URL })

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/chat/ws", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: got %d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/chat/ws?user_id=1", nil)
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{ID: "42", Token: "tok"}))
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Cookie", "s=1")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot || gotPath != "/ws" || gotQuery != "" || gotAuth != "Bearer tok" {
		t.Fatalf("code=%d path=%q query=%q auth=%q", rec.Code, gotPath, gotQuery, gotAuth)
	}
}
