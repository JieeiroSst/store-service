package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
	"github.com/JIeeiroSst/tool-service/internal/security"
	"github.com/JIeeiroSst/tool-service/internal/webqa"
)

type scopeItem struct {
	Area   string `json:"area"`
	Status string `json:"status"`
	Note   string `json:"note"`
}

func scopeMatrix() []scopeItem {
	return []scopeItem{
		{"API functional", "automated", "AI-proposed cases (trusted only when grounded, consistent and verified) plus the human-approved golden suite."},
		{"Regression", "automated", "Golden suite runs deterministically without AI."},
		{"Requirement coverage", "automated", "Statements of the document not quoted by any test are listed and get a second generation round."},
		{"Performance", "automated", "Latency percentiles and throughput; stress ramp to the breaking point and long soak runs with drift detection (background jobs)."},
		{"Security baseline (passive)", "automated", "TLS, headers, cookies, CORS, exposed files, anonymous access, error leakage."},
		{"Security, active", "automated", "Injection probes, malformed-input fuzzing, negative-money checks and an authorization matrix. Needs consent, an allow-list and an API token; writes need include_writes."},
		{"Exploratory testing", "automated", "Edge-case values submitted through the site's own forms."},
		{"Web pages", "automated", "Links, HTML quality, alt text, labels, headings."},
		{"Browser: responsive, contrast, keyboard, screen-reader tree, localization", "automated", "Real Chromium at desktop, tablet and mobile sizes and per locale; accessibility-tree checks stand in for a screen reader."},
		{"Visual regression", "automated", "Screenshots compared with a stored baseline; diff image saved."},
		{"Database and data integrity", "automated", "PostgreSQL and MySQL: keys, orphans, money columns and your SQL business rules, in read-only transactions."},
		{"Async jobs, queues, integrations", "automated", "Eventual-consistency and idempotency flows; RabbitMQ health through its management API."},
		{"Mobile apps: static analysis of APK and IPA", "automated", "Manifest, permissions, exported components, debug flags, signing, embedded secrets, ATS, usage descriptions versus APIs really used, provisioning. No device needed."},
		{"Mobile apps: Android on a device or emulator", "automated", "Through adb: install, cold start time, memory, jank, crashes/ANRs from logcat, monkey test, accessibility labels and touch targets, rotation, large font, dark mode, RTL, backgrounding."},
		{"Mobile apps: iOS on a simulator", "automated", "Through simctl on a macOS host with Xcode: launch, screenshots, crash reports, dark mode, largest Dynamic Type, RTL, deep links. Not available from the Linux image."},
		{"UI end-to-end and cross-browser", "generated", "Playwright script and a config for Chromium, Firefox and WebKit are generated for your CI; only Chromium runs inside this service."},
	}
}

func quotesOf(cases []apitest.TestCase) []string {
	var q []string
	for _, c := range cases {
		if c.SpecQuote != "" {
			q = append(q, c.SpecQuote)
		}
	}
	return q
}

func trustedQuotes(cases []apitest.TestCase, trust []caseTrust) []string {
	var q []string
	for i, c := range cases {
		if i < len(trust) && trust[i].Trusted && c.SpecQuote != "" {
			q = append(q, c.SpecQuote)
		}
	}
	return q
}

func firstN(s []string, n int) []string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func protectedPaths(p taskPlan) []string {
	var out []string
	for _, e := range p.Endpoints {
		auth := strings.ToLower(strings.TrimSpace(e.Auth))
		if auth == "" || auth == "none" || auth == "public" || auth == "no" {
			continue
		}
		if m := strings.ToUpper(e.Method); (m == "" || m == "GET") && !strings.ContainsAny(e.Path, "{}:") && e.Path != "" {
			out = append(out, e.Path)
		}
	}
	return out
}

func applyProductChecks(g *gate, sec []security.Finding, web *webqa.Report, blockUnreviewed bool) {
	if g == nil {
		return
	}
	var fail, review []string
	for _, f := range sec {
		switch f.Severity {
		case security.High:
			fail = append(fail, fmt.Sprintf("security %s: %s", f.ID, f.Title))
		case security.Medium:
			review = append(review, fmt.Sprintf("security %s: %s", f.ID, f.Title))
		}
	}
	if web != nil && web.Errors > 0 {
		fail = append(fail, fmt.Sprintf("web: %d error(s), %d broken link(s)", web.Errors, len(web.BrokenLinks)))
	}
	if len(fail) > 0 {
		g.Passed, g.Status = false, "failed"
		g.Reasons = append(g.Reasons, fail...)
	}
	if len(review) > 0 {
		g.NeedsReview = append(g.NeedsReview, review...)
		g.Reasons = append(g.Reasons, review...)
		if g.Status == "passed" {
			g.Status = "needs_review"
			g.Passed = !blockUnreviewed
		}
	}
}

func (h *Handler) qaScan(c *gin.Context) {
	var req struct {
		BaseURL        string   `json:"base_url" binding:"required"`
		ProtectedPaths []string `json:"protected_paths"`
		Security       *bool    `json:"security"`
		Crawl          *bool    `json:"crawl"`
		MaxPages       int      `json:"max_pages"`
		Unreviewed     string   `json:"unreviewed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.checkTarget(req.BaseURL); err != nil {
		badRequest(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	res := gin.H{"scope": scopeMatrix()}
	g := &gate{Passed: true, Status: "passed"}
	var sec []security.Finding
	var web *webqa.Report
	if req.Security == nil || *req.Security {
		sec = security.Scan(ctx, h.httpClient, req.BaseURL, security.Options{ProtectedPaths: req.ProtectedPaths})
		res["security"] = sec
	}
	if req.Crawl == nil || *req.Crawl {
		w := webqa.Crawl(ctx, h.httpClient, req.BaseURL, clamp(req.MaxPages, 1, 100, 20), 2)
		web = &w
		res["web"] = w
	}
	applyProductChecks(g, sec, web, req.Unreviewed != "warn")
	res["gate"] = g

	status := http.StatusOK
	if !g.Passed && c.Query("fail_http") == "true" {
		status = http.StatusUnprocessableEntity
	}
	c.JSON(status, res)
}

func (h *Handler) generateRuns(ctx context.Context, system, prompt string, n int) ([][]apitest.TestCase, error) {
	var runs [][]apitest.TestCase
	var last error
	for k := 0; k < n; k++ {
		temp := 0.2
		if k > 0 {
			temp = 0.7
		}
		var gen struct {
			Cases []apitest.TestCase `json:"cases"`
		}
		if err := h.ai.ChatJSONTemp(ctx, system, prompt, temp, &gen); err != nil {
			last = err
			continue
		}
		runs = append(runs, gen.Cases)
	}
	if len(runs) == 0 {
		return nil, last
	}
	return runs, nil
}

type e2eResult struct {
	Scenarios []struct {
		Name     string   `json:"name"`
		Steps    []string `json:"steps"`
		Expected string   `json:"expected"`
	} `json:"scenarios"`
	Script string `json:"playwright_script"`
	Config string `json:"playwright_config"`
	Note   string `json:"note"`
}

const e2ePrompt = `You are a QA automation engineer. From the task document, list the end-to-end user scenarios (UI flows) and write ONE
Playwright Test (TypeScript) file covering them. Use the baseURL from config (page.goto with relative paths), accessible locators
(getByRole/getByLabel/getByText), web-first assertions, and no hard-coded sleeps. Only use UI elements the document describes.
Also write playwright.config.ts with projects for chromium, firefox, webkit, "Mobile Chrome" (Pixel 7) and "Mobile Safari" (iPhone 14),
baseURL from process.env.BASE_URL, retries 1 on CI, and trace on first retry.
Reply ONLY with JSON: {"scenarios":[{"name":"","steps":[""],"expected":""}],"playwright_script":"<typescript source>","playwright_config":"<typescript source>"}`

func (h *Handler) generateE2E(ctx context.Context, planJSON []byte, doc string) *e2eResult {
	var out e2eResult
	if err := h.ai.ChatJSON(ctx, e2ePrompt, fmt.Sprintf("Plan:\n%s\n\nDocument:\n%s", planJSON, truncate(doc, 15000)), &out); err != nil {
		out.Note = "generation failed: " + err.Error()
		return &out
	}
	out.Note = "Generated by AI and NOT executed. Review it, then run it with Playwright in your own pipeline (npx playwright test) to cover Firefox, WebKit and mobile browsers."
	return &out
}

func checkRedirect(allowed map[string]bool) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("stopped after 5 redirects")
		}
		if len(allowed) == 0 {
			return nil
		}
		host := req.URL.Hostname()
		if u, err := url.Parse(req.URL.String()); err == nil {
			host = u.Hostname()
		}
		if !allowed[strings.ToLower(host)] {
			return fmt.Errorf("redirect to %q is not in TARGET_ALLOWLIST", host)
		}
		return nil
	}
}
