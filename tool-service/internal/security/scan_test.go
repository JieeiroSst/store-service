package security

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ids(fs []Finding) map[string]string {
	m := map[string]string{}
	for _, f := range fs {
		m[f.ID] = f.Severity
	}
	return m
}

func insecure() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "TRACE" {
			w.Write([]byte("TRACE / HTTP/1.1\nX-Trace-Probe: " + r.Header.Get("X-Trace-Probe")))
			return
		}
		if o := r.Header.Get("Origin"); o != "" {
			w.Header().Set("Access-Control-Allow-Origin", o)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Server", "nginx/1.14.0")
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "1"})
		if r.URL.Path == "/" {
			w.Write([]byte("<html>home</html>"))
			return
		}
		if r.URL.Query().Get("q") != "" {
			w.WriteHeader(500)
			w.Write([]byte("panic: runtime error\ngoroutine 1 [running]"))
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/.env", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("DB_PASSWORD=secret\n")) })
	mux.HandleFunc("/admin/users", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`[{"id":1}]`)) })
	return mux
}

func TestScanFindsProblems(t *testing.T) {
	srv := httptest.NewServer(insecure())
	defer srv.Close()
	got := ids(Scan(context.Background(), srv.Client(), srv.URL, Options{ProtectedPaths: []string{"/admin/users"}}))
	for id, sev := range map[string]string{
		"SEC-HDR-003": Medium, "SEC-HDR-004": Medium, "SEC-CORS-001": High, "SEC-MTH-001": Medium,
		"SEC-EXP-ENV": High, "SEC-AUTH-001": High, "SEC-ERR-001": Medium, "SEC-CK-001": Medium, "SEC-HDR-006": Low,
	} {
		if got[id] != sev {
			t.Errorf("%s: want %s, got %q (all: %v)", id, sev, got[id], got)
		}
	}
}

func TestScanHardenedIsQuiet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-RateLimit-Limit", "100")
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()
	for _, f := range Scan(context.Background(), srv.Client(), srv.URL, Options{}) {
		if f.Severity != Info {
			t.Errorf("hardened server should only yield info findings, got %+v", f)
		}
	}
}

func TestScanSoftNotFoundNoFalsePositive(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html>app DB=1 [core] ref: swagger # HELP phpinfo()</html>"))
	}))
	defer srv.Close()
	for _, f := range Scan(context.Background(), srv.Client(), srv.URL, Options{ProtectedPaths: []string{"/api/x"}}) {
		if strings.HasPrefix(f.ID, "SEC-EXP") || f.ID == "SEC-AUTH-001" {
			t.Errorf("catch-all 200 must not be reported: %+v", f)
		}
	}
}

func TestScanUnreachable(t *testing.T) {
	fs := Scan(context.Background(), http.DefaultClient, "http://127.0.0.1:1", Options{})
	if len(fs) != 1 || fs[0].Severity != High {
		t.Fatalf("got %+v", fs)
	}
}
