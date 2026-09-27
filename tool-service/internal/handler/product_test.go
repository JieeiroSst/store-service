package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/tool-service/internal/learning"
	"github.com/JIeeiroSst/tool-service/internal/ollama"
	"github.com/JIeeiroSst/tool-service/internal/security"
	"github.com/JIeeiroSst/tool-service/internal/webqa"
)

func TestApplyProductChecks(t *testing.T) {
	high := []security.Finding{{ID: "SEC-X", Severity: security.High, Title: "leak"}}
	med := []security.Finding{{ID: "SEC-Y", Severity: security.Medium, Title: "no csp"}}
	low := []security.Finding{{ID: "SEC-Z", Severity: security.Low, Title: "nit"}, {ID: "SEC-I", Severity: security.Info}}

	cases := []struct {
		name       string
		sec        []security.Finding
		web        *webqa.Report
		block      bool
		wantStatus string
		wantPassed bool
	}{
		{"clean", low, &webqa.Report{Warnings: 5}, true, "passed", true},
		{"high security fails", high, nil, true, "failed", false},
		{"broken link fails", nil, &webqa.Report{Errors: 1, BrokenLinks: []webqa.BrokenLink{{URL: "/x"}}}, true, "failed", false},
		{"medium needs review (blocking)", med, nil, true, "needs_review", false},
		{"medium needs review (warn only)", med, nil, false, "needs_review", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := &gate{Passed: true, Status: "passed"}
			applyProductChecks(g, tc.sec, tc.web, tc.block)
			if g.Status != tc.wantStatus || g.Passed != tc.wantPassed {
				t.Fatalf("status=%s passed=%v, want %s/%v", g.Status, g.Passed, tc.wantStatus, tc.wantPassed)
			}
		})
	}

	g := &gate{Passed: false, Status: "failed"}
	applyProductChecks(g, med, nil, false)
	if g.Status != "failed" || g.Passed {
		t.Fatalf("failed gate was downgraded: %+v", g)
	}
	applyProductChecks(nil, high, nil, true)
}

func TestProtectedPaths(t *testing.T) {
	var p taskPlan
	add := func(m, path, auth string) {
		p.Endpoints = append(p.Endpoints, struct {
			Method      string `json:"method"`
			Path        string `json:"path"`
			Description string `json:"description"`
			Auth        string `json:"auth"`
		}{Method: m, Path: path, Auth: auth})
	}
	add("GET", "/orders/1", "bearer")
	add("GET", "/health", "none")
	add("POST", "/orders", "bearer")
	add("GET", "/users/{id}", "bearer")
	add("", "/me", "jwt")
	got := protectedPaths(p)
	if len(got) != 2 || got[0] != "/orders/1" || got[1] != "/me" {
		t.Fatalf("got %v", got)
	}
}

func TestRedirectCannotEscapeAllowList(t *testing.T) {
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("secret")) }))
	defer evil.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL, http.StatusFound)
	}))
	defer good.Close()

	mem, err := learning.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ai := ollama.NewClient("http://unused", "m", 0)
	h := New(ai, mem, LearnConfig{}, []string{"allowed.example"})
	if _, err := h.httpClient.Get(good.URL); err == nil {
		t.Fatal("redirect to a host outside the allow-list must be refused")
	}
	open := New(ai, mem, LearnConfig{}, nil)
	resp, err := open.httpClient.Get(good.URL)
	if err != nil {
		t.Fatalf("without allow-list redirects work: %v", err)
	}
	resp.Body.Close()
}
