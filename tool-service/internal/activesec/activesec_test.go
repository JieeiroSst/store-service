package activesec

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func vulnerable() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html>results for %s</html>", r.URL.Query().Get("q"))
	})
	mux.HandleFunc("/item", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("id"), "'") {
			w.WriteHeader(500)
			w.Write([]byte("You have an error in your SQL syntax near '''"))
			return
		}
		w.Write([]byte("item"))
	})
	mux.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("file"), "etc/passwd") {
			w.Write([]byte("root:x:0:0:root:/root:/bin/sh"))
			return
		}
		w.Write([]byte("page"))
	})
	mux.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
		n := r.URL.Query().Get("name")
		if n == "{{7*7}}" || n == "${7*7}" {
			n = "49"
		}
		w.Write([]byte("hello " + n))
	})
	mux.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {
		var b map[string]any
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			w.WriteHeader(400)
			return
		}
		if _, isStr := b["amount"].(string); isStr {
			w.WriteHeader(500)
			w.Write([]byte("panic: interface conversion\ngoroutine 7 [running]:"))
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		tok := r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/orders/1":
			w.Write([]byte(`{"id":1}`))
		case "/orders/2":
			if tok == "bob" {
				w.Write([]byte(`{"id":2}`))
				return
			}
			w.WriteHeader(403)
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/find", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("term"), "'") {
			w.WriteHeader(500)
			return
		}
		w.Write([]byte("ok"))
	})
	return mux
}

func hardened() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html>results for %s</html>", html.EscapeString(r.URL.Query().Get("q")))
	})
	mux.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Amount float64 `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil || b.Amount <= 0 || b.Amount > 1e9 {
			w.WriteHeader(422)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "alice" {
			w.WriteHeader(403)
			return
		}
		w.Write([]byte(`{"id":1}`))
	})
	return mux
}

func ids(r Result) map[string]bool {
	m := map[string]bool{}
	for _, f := range r.Findings {
		m[f.ID] = true
	}
	return m
}

func TestVulnerableAppIsCaught(t *testing.T) {
	srv := httptest.NewServer(vulnerable())
	defer srv.Close()
	res := Run(context.Background(), srv.Client(), srv.URL, Options{
		Endpoints: []Endpoint{
			{Path: "/search", Params: map[string]string{"q": "x"}},
			{Path: "/item", Params: map[string]string{"id": "1"}},
			{Path: "/view", Params: map[string]string{"file": "a.txt"}},
			{Path: "/greet", Params: map[string]string{"name": "bob"}},
			{Method: "POST", Path: "/pay", Body: map[string]any{"amount": 10.5, "note": "x"}},
		},
		Identities:    []Identity{{Name: "alice", Headers: map[string]string{"Authorization": "alice"}}, {Name: "bob", Headers: map[string]string{"Authorization": "bob"}}},
		Resources:     []Resource{{Path: "/orders/1", Owner: "alice"}, {Path: "/orders/2", Owner: "bob"}},
		Forms:         []Form{{Method: "GET", Action: srv.URL + "/find", Fields: []FormField{{Name: "term", Type: "text"}}}},
		IncludeWrites: true,
	})
	got := ids(res)
	for _, want := range []string{"ACT-XSS-001", "ACT-SQLI-001", "ACT-TRAV-001", "ACT-SSTI-001", "ACT-MONEY-001", "ACT-MONEY-002", "ACT-MONEY-003",
		"ACT-FUZZ-001", "ACT-ERR-001", "ACT-AUTHZ-001", "ACT-FORM-001"} {
		if !got[want] {
			t.Errorf("missing %s; got %v", want, got)
		}
	}

	authz := 0
	for _, f := range res.Findings {
		if f.ID == "ACT-AUTHZ-001" {
			authz++
			if !strings.Contains(f.Endpoint, "/orders/1") {
				t.Errorf("protected resource wrongly flagged: %+v", f)
			}
		}
	}
	if authz != 2 {
		t.Errorf("want 2 authz findings (bob + anonymous on /orders/1), got %d", authz)
	}
}

func TestHardenedAppIsQuiet(t *testing.T) {
	srv := httptest.NewServer(hardened())
	defer srv.Close()
	res := Run(context.Background(), srv.Client(), srv.URL, Options{
		Endpoints:     []Endpoint{{Path: "/search", Params: map[string]string{"q": "x"}}, {Method: "POST", Path: "/pay", Body: map[string]any{"amount": 10.5}}},
		Identities:    []Identity{{Name: "alice", Headers: map[string]string{"Authorization": "alice"}}, {Name: "bob", Headers: map[string]string{"Authorization": "bob"}}},
		Resources:     []Resource{{Path: "/orders/1", Owner: "alice"}},
		IncludeWrites: true,
	})
	if len(res.Findings) != 0 {
		t.Fatalf("hardened app must produce no findings, got %+v", res.Findings)
	}
}

func TestWritesRequireOptIn(t *testing.T) {
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			posts++
		}
	}))
	defer srv.Close()
	res := Run(context.Background(), srv.Client(), srv.URL, Options{
		Endpoints: []Endpoint{{Method: "POST", Path: "/pay", Body: map[string]any{"amount": 1}}},
		Forms:     []Form{{Method: "POST", Action: srv.URL + "/f", Fields: []FormField{{Name: "a", Type: "text"}}}},
	})
	if posts != 0 || res.Requests != 0 || len(res.Skipped) != 2 {
		t.Fatalf("state-changing requests were sent without include_writes: posts=%d res=%+v", posts, res)
	}
}

func TestRequestBudget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	res := Run(context.Background(), srv.Client(), srv.URL, Options{
		Endpoints:     []Endpoint{{Method: "POST", Path: "/pay", Body: map[string]any{"amount": 1, "b": 2, "c": 3}}},
		IncludeWrites: true, MaxRequests: 25,
	})
	if res.Requests != 25 {
		t.Fatalf("budget not enforced: sent %d", res.Requests)
	}
	if len(res.Skipped) == 0 || !strings.Contains(res.Skipped[len(res.Skipped)-1], "budget") {
		t.Fatalf("truncation must be reported: %+v", res.Skipped)
	}
}
