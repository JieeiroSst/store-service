package http

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func logged(t *testing.T, statusFor func(i int) int, path string, method string, n int) int {
	t.Helper()
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	i := 0
	h := logging(log, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusFor(i))
		i++
	}))
	for k := 0; k < n; k++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, nil))
	}
	if buf.Len() == 0 {
		return 0
	}
	return strings.Count(buf.String(), "\n")
}

// Under a rush almost every request is a cheap refusal. Logging each one costs more CPU than answering it, and a
// process starved by its own logging misses its health probes and gets restarted by the platform.
func TestARushOfRefusalsDoesNotWriteALogLinePerRequest(t *testing.T) {
	const n = 10_000
	if got := logged(t, func(int) int { return http.StatusConflict }, "/api/v1/reservations", http.MethodPost, n); got > n/50 || got == 0 {
		t.Errorf("%d refusals wrote %d lines; want a small sample, not none and not all", n, got)
	}
	if got := logged(t, func(int) int { return http.StatusTooManyRequests }, "/api/v1/reservations", http.MethodPost, n); got > n/50 {
		t.Errorf("429s wrote %d lines", got)
	}
	if got := logged(t, func(int) int { return http.StatusOK }, "/api/v1/hotels", http.MethodGet, n); got > n/50 {
		t.Errorf("plain reads wrote %d lines", got)
	}
}

func TestErrorsChangesAndSlowRequestsAreAlwaysLogged(t *testing.T) {
	if got := logged(t, func(int) int { return http.StatusInternalServerError }, "/api/v1/hotels", http.MethodGet, 50); got != 50 {
		t.Errorf("every 5xx must be logged: %d of 50", got)
	}
	if got := logged(t, func(int) int { return http.StatusCreated }, "/api/v1/reservations", http.MethodPost, 50); got != 50 {
		t.Errorf("every change that succeeded must be logged: %d of 50", got)
	}
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	slow := logging(log, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(600 * time.Millisecond) }))
	slow.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/hotels", nil))
	if !strings.Contains(buf.String(), "slow") {
		t.Errorf("a slow request must be logged: %q", buf.String())
	}
}

func TestHealthChecksAreNeverLogged(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz"} {
		if got := logged(t, func(int) int { return http.StatusOK }, path, http.MethodGet, 1_000); got != 0 {
			t.Errorf("%s wrote %d lines", path, got)
		}
	}
}

func TestInFlightLimitTurnsExcessRequestsAwayButNeverHealthChecks(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{}, 4)
	h := limitInFlight(2, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release
	}))
	done := make(chan struct{}, 2)
	for i := 0; i < 2; i++ { // fill both slots
		go func() {
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/reservations", nil))
			done <- struct{}{}
		}()
	}
	<-entered
	<-entered
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/reservations", nil))
	if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("the third request must be refused at once with Retry-After: %d %v", rec.Code, rec.Header())
	}
	// A health check must still be served while the pod is full, or the platform would think it dead.
	hc := httptest.NewRecorder()
	hDone := make(chan struct{})
	go func() {
		h.ServeHTTP(hc, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		close(hDone)
	}()
	select {
	case <-entered: // the health check reached the handler
	case <-time.After(time.Second):
		t.Fatal("the health check was turned away")
	}
	close(release)
	<-hDone
	<-done
	<-done
}
