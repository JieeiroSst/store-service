package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
	"github.com/JIeeiroSst/tool-service/internal/benchmark"
	"github.com/JIeeiroSst/tool-service/internal/jobs"
	"github.com/JIeeiroSst/tool-service/internal/learning"
	"github.com/JIeeiroSst/tool-service/internal/ollama"
	"github.com/JIeeiroSst/tool-service/internal/suite"
)

type Handler struct {
	ai  *ollama.Client
	mem *learning.Store
	cfg LearnConfig

	suites           *suite.Store
	infra            Infra
	jobs             *jobs.Manager
	approvedAtEvolve int
	apiToken         string

	allowedHosts map[string]bool
	httpClient   *http.Client
}

func New(ai *ollama.Client, mem *learning.Store, cfg LearnConfig, allowedHosts []string) *Handler {
	allowed := map[string]bool{}
	for _, h := range allowedHosts {
		if h = strings.TrimSpace(h); h != "" {
			allowed[strings.ToLower(h)] = true
		}
	}
	if m := mem.ActiveModel(); m != "" {
		ai.SetModel(m)
	}
	return &Handler{
		ai:           ai,
		mem:          mem,
		cfg:          cfg,
		jobs:         jobs.New(),
		infra:        Infra{DataDir: "./data"},
		allowedHosts: allowed,
		httpClient: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: checkRedirect(allowed),
			Transport: &http.Transport{
				MaxIdleConnsPerHost: benchmark.MaxConcurrency,
				MaxIdleConns:        benchmark.MaxConcurrency,
			},
		},
	}
}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "model": h.ai.Model()}) })
	v1 := r.Group("/api/v1")
	if h.apiToken != "" {
		v1.Use(h.requireToken)
	}
	v1.POST("/testcases/generate", h.generateTestCases)
	v1.POST("/testcases/run", h.runTestCases)
	v1.POST("/benchmark", h.runBenchmark)
	v1.POST("/analyze/code", h.analyzeCode)
	v1.POST("/tasks/analyze", h.analyzeTask)
	v1.POST("/qa/scan", h.qaScan)
	v1.POST("/export/xlsx", h.exportXLSX)
	h.registerLearning(v1)
	h.registerSuites(v1)
	h.registerAdvanced(v1)
}

func (h *Handler) checkTarget(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("invalid target url %q", raw)
	}
	if len(h.allowedHosts) == 0 {
		return nil
	}
	host := u.Host
	if hn, _, err := net.SplitHostPort(u.Host); err == nil {
		host = hn
	}
	if !h.allowedHosts[strings.ToLower(host)] {
		return fmt.Errorf("host %q is not in TARGET_ALLOWLIST", host)
	}
	return nil
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func aiError(c *gin.Context, err error) {
	c.JSON(http.StatusBadGateway, gin.H{"error": "ollama: " + err.Error()})
}

type generateRequest struct {
	Spec    string `json:"spec" binding:"required"`
	BaseURL string `json:"base_url"`
	Count   int    `json:"count"`

	Run bool `json:"run"`
}

const testCaseSystemPrompt = `You are a senior QA engineer. Given an API specification or handler code, design API test cases
covering: happy path, validation errors, missing/invalid auth, boundary values, not-found, and idempotency.
Reply ONLY with JSON of the form:
{"cases":[{"name":"","method":"GET","path":"/x","headers":{},"body":null,"expect_status":200,
"expect_body_contains":[],"expect_json":{"data.id":1},"max_latency_ms":500}]}
Use realistic data. Do not include prose.`

func (h *Handler) generateTestCases(c *gin.Context) {
	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if req.Count <= 0 || req.Count > 30 {
		req.Count = 10
	}
	if req.Run {
		if err := h.checkTarget(req.BaseURL); err != nil {
			badRequest(c, err)
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	var out struct {
		Cases []apitest.TestCase `json:"cases"`
	}
	prompt := fmt.Sprintf("Generate about %d test cases for:\n\n%s", req.Count, req.Spec)
	system, emb := h.recall(ctx, learning.KindTestCases, req.Spec, testCaseSystemPrompt)
	if err := h.ai.ChatJSON(ctx, system, prompt, &out); err != nil {
		aiError(c, err)
		return
	}
	resp := gin.H{"model": h.ai.Model(), "cases": out.Cases}
	signals := map[string]any{}
	if req.Run {
		report := apitest.Run(ctx, h.httpClient, req.BaseURL, out.Cases)
		resp["report"] = report
		signals = reportSignals(report)
	}
	resp["experience_id"] = h.record(learning.KindTestCases, req.Spec, out, emb, signals)
	c.JSON(http.StatusOK, resp)
}

type runRequest struct {
	BaseURL string             `json:"base_url" binding:"required"`
	Cases   []apitest.TestCase `json:"cases" binding:"required"`

	Analyze bool `json:"analyze"`
}

func (h *Handler) runTestCases(c *gin.Context) {
	var req runRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.checkTarget(req.BaseURL); err != nil {
		badRequest(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	report := apitest.Run(ctx, h.httpClient, req.BaseURL, req.Cases)
	resp := gin.H{"report": report}
	if req.Analyze && report.Failed > 0 {
		resp["analysis"] = h.explain(ctx, "Explain the likely root causes of these failed API tests and how to fix them.", report)
	}
	c.JSON(http.StatusOK, resp)
}

type benchmarkRequest struct {
	benchmark.Config
	Analyze bool `json:"analyze"`
}

func (h *Handler) runBenchmark(c *gin.Context) {
	var req benchmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := req.Config.Validate(); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.checkTarget(req.URL); err != nil {
		badRequest(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), benchmark.MaxDuration+time.Minute)
	defer cancel()

	stats := benchmark.Run(ctx, h.httpClient, req.Config)
	resp := gin.H{"stats": stats}
	if req.Analyze {
		resp["analysis"] = h.explain(ctx,
			"Analyze this HTTP load-test result: identify bottlenecks, tail-latency problems, error patterns and give concrete tuning suggestions.", stats)
	}
	c.JSON(http.StatusOK, resp)
}

type codeRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language"`

	Focus string `json:"focus"`
}

const codeSystemPrompt = `You are a principal engineer reviewing code. Reply ONLY with JSON:
{"summary":"","logic_issues":[{"line":"","issue":"","fix":""}],
"performance":{"time_complexity":"","space_complexity":"","hotspots":[""],"optimizations":[""]},
"unit_tests":"<runnable test code as a string>","benchmark_code":"<runnable benchmark code as a string>"}
Be concrete and only report issues actually present in the code.`

func (h *Handler) analyzeCode(c *gin.Context) {
	var req codeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if req.Language == "" {
		req.Language = "go"
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	prompt := fmt.Sprintf("Language: %s\nFocus: %s\n\n```\n%s\n```", req.Language, orDefault(req.Focus, "all"), req.Code)
	var out map[string]any
	system, emb := h.recall(ctx, learning.KindCode, req.Code, codeSystemPrompt)
	if err := h.ai.ChatJSON(ctx, system, prompt, &out); err != nil {
		aiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"model": h.ai.Model(), "analysis": out,
		"experience_id": h.record(learning.KindCode, req.Code, out, emb, nil),
	})
}

func (h *Handler) explain(ctx context.Context, instruction string, data any) any {
	b, _ := json.Marshal(data)
	text, err := h.ai.Chat(ctx, "You are a performance and QA expert. Be concise and actionable.", instruction+"\n\n"+string(b), false)
	if err != nil {
		return gin.H{"error": err.Error()}
	}
	return text
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func (h *Handler) SetSuites(s *suite.Store) { h.suites = s }

func (h *Handler) SetAPIToken(t string) { h.apiToken = t }

func (h *Handler) requireToken(c *gin.Context) {
	got := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if got == "" {
		got = c.GetHeader("X-API-Key")
	}
	if subtle.ConstantTimeCompare([]byte(got), []byte(h.apiToken)) != 1 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid API token"})
		return
	}
	c.Next()
}
