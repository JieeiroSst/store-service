package probe

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

func TestHTTPStatusMapping(t *testing.T) {
	var gotMethod, gotBody, gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotHeader = r.Method, r.Header.Get("X-Probe")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		code := map[string]int{"/ok": 200, "/busy": 429, "/err": 500}[r.URL.Path]
		w.WriteHeader(code)
		_, _ = io.WriteString(w, "body")
	}))
	defer srv.Close()

	p := NewProber()
	cases := map[string]domain.HealthStatus{
		"/ok":   domain.HealthPassing,
		"/busy": domain.HealthWarning,
		"/err":  domain.HealthCritical,
	}
	for path, want := range cases {
		st, out := p.Probe(context.Background(), domain.Check{Type: domain.CheckHTTP, HTTP: srv.URL + path})
		if st != want {
			t.Errorf("%s: status %s, want %s (%s)", path, st, want, out)
		}
		if !strings.Contains(out, "Output: body") {
			t.Errorf("%s: output %q", path, out)
		}
	}

	_, _ = p.Probe(context.Background(), domain.Check{
		Type: domain.CheckHTTP, HTTP: srv.URL + "/ok", Method: "POST", Body: "ping",
		Header: map[string][]string{"X-Probe": {"1"}},
	})
	if gotMethod != "POST" || gotBody != "ping" || gotHeader != "1" {
		t.Fatalf("request = %s %q header %q", gotMethod, gotBody, gotHeader)
	}
}

func TestHTTPTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if st, _ := NewProber().Probe(ctx, domain.Check{Type: domain.CheckHTTP, HTTP: srv.URL}); st != domain.HealthCritical {
		t.Fatalf("status = %s, want critical", st)
	}
}

func TestTCP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	p := NewProber()
	if st, out := p.Probe(context.Background(), domain.Check{Type: domain.CheckTCP, TCP: addr}); st != domain.HealthPassing {
		t.Fatalf("open port: %s %s", st, out)
	}
	ln.Close()
	if st, _ := p.Probe(context.Background(), domain.Check{Type: domain.CheckTCP, TCP: addr}); st != domain.HealthCritical {
		t.Fatalf("closed port: %s", st)
	}
}
