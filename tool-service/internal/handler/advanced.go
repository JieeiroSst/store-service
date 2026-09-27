package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/activesec"
	"github.com/JIeeiroSst/tool-service/internal/asyncflow"
	"github.com/JIeeiroSst/tool-service/internal/benchmark"
	"github.com/JIeeiroSst/tool-service/internal/browserqa"
	"github.com/JIeeiroSst/tool-service/internal/dbcheck"
	"github.com/JIeeiroSst/tool-service/internal/jobs"
	"github.com/JIeeiroSst/tool-service/internal/queuecheck"
	"github.com/JIeeiroSst/tool-service/internal/webqa"
)

type DBConn struct {
	Dialect string `json:"dialect"`
	DSN     string `json:"dsn"`
}

type Infra struct {
	DBs       map[string]DBConn
	Rabbit    map[string]queuecheck.RabbitMQ
	DataDir   string
	NoSandbox bool
}

func (h *Handler) SetInfra(i Infra) {
	h.infra = i
	h.jobs = jobs.New()
}

type inputError struct{ error }

func badInput(f string, a ...any) error { return inputError{fmt.Errorf(f, a...)} }

func (h *Handler) runQA(c *gin.Context, kind string, timeout time.Duration, fn func(ctx context.Context, progress func(string)) (gin.H, error)) {
	if c.Query("async") == "true" {
		j := h.jobs.Start(kind, timeout, func(ctx context.Context, p func(string)) (any, error) { return fn(ctx, p) })
		c.JSON(http.StatusAccepted, gin.H{"job_id": j.ID, "status_url": "/api/v1/jobs/" + j.ID})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	out, err := fn(ctx, nil)
	var ie inputError
	switch {
	case errors.As(err, &ie):
		badRequest(c, err)
		return
	case err != nil:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	status := http.StatusOK
	if g, ok := out["gate"].(*gate); ok && !g.Passed && c.Query("fail_http") == "true" {
		status = http.StatusUnprocessableEntity
	}
	c.JSON(status, out)
}

func (h *Handler) registerAdvanced(v1 *gin.RouterGroup) {
	q := v1.Group("/qa")
	q.POST("/database", h.qaDatabase)
	q.POST("/async", h.qaAsync)
	q.POST("/queue", h.qaQueue)
	q.POST("/browser", h.qaBrowser)
	q.POST("/active", h.qaActive)
	h.registerMobile(q)
	v1.POST("/jobs/stress", h.jobStress)
	v1.POST("/jobs/soak", h.jobSoak)
	v1.GET("/jobs", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"jobs": h.jobs.List()}) })
	v1.GET("/jobs/:id", func(c *gin.Context) {
		j, ok := h.jobs.Get(c.Param("id"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found (jobs are kept in memory and lost on restart)"})
			return
		}
		c.JSON(http.StatusOK, j)
	})
	v1.POST("/jobs/:id/cancel", func(c *gin.Context) {
		if !h.jobs.Cancel(c.Param("id")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no running job with that id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"canceled": true})
	})
	v1.GET("/artifacts/:file", h.artifact)
}

func (h *Handler) needsToken(c *gin.Context) bool {
	if h.apiToken == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "this endpoint requires API_TOKEN to be configured on the server"})
		return false
	}
	return true
}

func combine(g *gate, blockUnreviewed bool, fail, review []string) *gate {
	if g == nil {
		g = &gate{Passed: true, Status: "passed"}
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
	return g
}

func (h *Handler) qaDatabase(c *gin.Context) {
	if !h.needsToken(c) {
		return
	}
	var req struct {
		Connection string              `json:"connection" binding:"required"`
		Assertions []dbcheck.Assertion `json:"assertions"`
		Unreviewed string              `json:"unreviewed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	conn, ok := h.infra.DBs[req.Connection]
	if !ok {
		var names []string
		for n := range h.infra.DBs {
			names = append(names, n)
		}
		badRequest(c, fmt.Errorf("unknown connection %q; configured: %v (set DB_CONNECTIONS on the server)", req.Connection, names))
		return
	}
	h.runQA(c, "database", 3*time.Minute, func(ctx context.Context, _ func(string)) (gin.H, error) {
		db, err := dbcheck.Open(conn.Dialect, conn.DSN)
		if err != nil {
			return nil, badInput("%v", err)
		}
		defer db.Close()
		rep, err := dbcheck.Audit(ctx, db, conn.Dialect, req.Assertions)
		if err != nil {
			return nil, err
		}
		var fail, review []string
		for _, f := range rep.Findings {
			switch f.Severity {
			case dbcheck.High:
				fail = append(fail, "database "+f.ID+": "+f.Title)
			case dbcheck.Medium:
				review = append(review, "database "+f.ID+": "+f.Title)
			}
		}
		for _, a := range rep.Assertions {
			if !a.Passed {
				fail = append(fail, fmt.Sprintf("assertion %q failed: %s", a.Name, a.Detail))
			}
		}
		return gin.H{"report": rep, "gate": combine(nil, req.Unreviewed != "warn", fail, review)}, nil
	})
}

func (h *Handler) qaAsync(c *gin.Context) {
	var req struct {
		BaseURL string         `json:"base_url" binding:"required"`
		Flow    asyncflow.Spec `json:"flow"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.checkTarget(req.BaseURL); err != nil {
		badRequest(c, err)
		return
	}
	timeout := time.Duration(clampInt(req.Flow.TimeoutSec, 1, 300, 30))*time.Second + time.Minute
	h.runQA(c, "async-flow", timeout, func(ctx context.Context, _ func(string)) (gin.H, error) {
		r := asyncflow.Run(ctx, h.httpClient, req.BaseURL, req.Flow)
		var fail []string
		for _, f := range r.Failures {
			fail = append(fail, "async flow: "+f)
		}
		return gin.H{"result": r, "gate": combine(nil, true, fail, nil)}, nil
	})
}

func (h *Handler) qaQueue(c *gin.Context) {
	if !h.needsToken(c) {
		return
	}
	var req struct {
		Broker     string `json:"broker" binding:"required"`
		MaxReady   int    `json:"max_ready"`
		MaxUnacked int    `json:"max_unacked"`
		Unreviewed string `json:"unreviewed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	cfg, ok := h.infra.Rabbit[req.Broker]
	if !ok {
		badRequest(c, fmt.Errorf("unknown broker %q (set RABBITMQ_CONNECTIONS on the server)", req.Broker))
		return
	}
	h.runQA(c, "queue", time.Minute, func(ctx context.Context, _ func(string)) (gin.H, error) {
		fs, err := queuecheck.CheckRabbitMQ(ctx, h.httpClient, cfg, queuecheck.Options{MaxReady: req.MaxReady, MaxUnacked: req.MaxUnacked})
		if err != nil {
			return nil, err
		}
		var fail, review []string
		for _, f := range fs {
			switch f.Severity {
			case "high":
				fail = append(fail, "queue "+f.ID+": "+f.Title)
			case "medium":
				review = append(review, "queue "+f.ID+": "+f.Title)
			}
		}
		return gin.H{"findings": fs, "gate": combine(nil, req.Unreviewed != "warn", fail, review)}, nil
	})
}

var hostRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func (h *Handler) qaBrowser(c *gin.Context) {
	var req struct {
		BaseURL        string   `json:"base_url" binding:"required"`
		URLs           []string `json:"urls"`
		MaxPages       int      `json:"max_pages"`
		Viewports      []string `json:"viewports"`
		Locales        []string `json:"locales"`
		Name           string   `json:"name"`
		UpdateBaseline bool     `json:"update_baseline"`
		DiffThreshold  float64  `json:"diff_threshold_pct"`
		StrictWarnings bool     `json:"strict_warnings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.checkTarget(req.BaseURL); err != nil {
		badRequest(c, err)
		return
	}
	for _, u := range req.URLs {
		if err := h.checkTarget(u); err != nil {
			badRequest(c, err)
			return
		}
	}
	if _, err := browserqa.FindChrome(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	maxPages := clampInt(req.MaxPages, 1, 10, 3)
	h.runQA(c, "browser", 20*time.Minute, func(ctx context.Context, progress func(string)) (gin.H, error) {
		urls := req.URLs
		if len(urls) == 0 {
			cr := webqa.Crawl(ctx, h.httpClient, req.BaseURL, maxPages, 1)
			for _, p := range cr.Pages {
				if p.Status > 0 && p.Status < 400 {
					urls = append(urls, p.URL)
				}
			}
			if len(urls) == 0 {
				urls = []string{req.BaseURL}
			}
		}
		if len(urls) > 10 {
			urls = urls[:10]
		}
		name := req.Name
		if name == "" {
			name = "default"
		}
		opt := browserqa.Options{
			Viewports:        pickViewports(req.Viewports),
			Locales:          req.Locales,
			Name:             hostRe.ReplaceAllString(name, "_"),
			BaselineDir:      filepath.Join(h.infra.DataDir, "baselines"),
			ArtifactDir:      filepath.Join(h.infra.DataDir, "artifacts"),
			UpdateBaseline:   req.UpdateBaseline,
			DiffThresholdPct: req.DiffThreshold,
			AllowedHosts:     h.allowedHosts,
			NoSandbox:        h.infra.NoSandbox,
			Progress:         progress,
		}
		rep, err := browserqa.Audit(ctx, urls, opt)
		if err != nil {
			return nil, err
		}
		var fail, review []string
		for _, r := range rep.Results {
			for _, is := range r.Issues {
				line := fmt.Sprintf("browser %s [%s%s] %s: %s", r.Page, r.Viewport, localeSuffix(r.Locale), is.Rule, is.Detail)
				switch {
				case is.Severity == browserqa.Error:
					fail = append(fail, line)
				case is.Severity == browserqa.Warn && req.StrictWarnings:
					review = append(review, line)
				}
			}
		}
		return gin.H{"report": rep, "gate": combine(nil, true, capList(fail, 20), capList(review, 20))}, nil
	})
}

func localeSuffix(l string) string {
	if l == "" {
		return ""
	}
	return " " + l
}

func pickViewports(names []string) []browserqa.Viewport {
	if len(names) == 0 {
		return browserqa.DefaultViewports
	}
	var out []browserqa.Viewport
	for _, n := range names {
		for _, v := range browserqa.DefaultViewports {
			if strings.EqualFold(v.Name, n) {
				out = append(out, v)
			}
		}
	}
	if len(out) == 0 {
		return browserqa.DefaultViewports
	}
	return out
}

func capList(l []string, n int) []string {
	if len(l) > n {
		return append(l[:n:n], fmt.Sprintf("... and %d more", len(l)-n))
	}
	return l
}

var artifactName = regexp.MustCompile(`^[a-zA-Z0-9._-]+\.png$`)

func (h *Handler) artifact(c *gin.Context) {
	name := c.Param("file")
	if !artifactName.MatchString(name) || strings.Contains(name, "..") {
		badRequest(c, fmt.Errorf("invalid artifact name"))
		return
	}
	p := filepath.Join(h.infra.DataDir, "artifacts", name)
	if _, err := os.Stat(p); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact not found"})
		return
	}
	c.File(p)
}

func (h *Handler) qaActive(c *gin.Context) {
	if !h.needsToken(c) {
		return
	}
	var req struct {
		BaseURL string `json:"base_url" binding:"required"`

		Active        bool                 `json:"active"`
		Endpoints     []activesec.Endpoint `json:"endpoints"`
		Identities    []activesec.Identity `json:"identities"`
		Resources     []activesec.Resource `json:"resources"`
		IncludeWrites bool                 `json:"include_writes"`
		ExploreForms  bool                 `json:"explore_forms"`
		MaxRequests   int                  `json:"max_requests"`
		Unreviewed    string               `json:"unreviewed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if !req.Active {
		badRequest(c, fmt.Errorf(`set "active": true to confirm you own the target and consent to injection, fuzzing and authorization probes`))
		return
	}
	if len(h.allowedHosts) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "active testing requires TARGET_ALLOWLIST to be set on the server"})
		return
	}
	if err := h.checkTarget(req.BaseURL); err != nil {
		badRequest(c, err)
		return
	}
	h.runQA(c, "active-security", 15*time.Minute, func(ctx context.Context, progress func(string)) (gin.H, error) {
		opt := activesec.Options{Endpoints: req.Endpoints, Identities: req.Identities, Resources: req.Resources,
			IncludeWrites: req.IncludeWrites, MaxRequests: req.MaxRequests}
		if req.ExploreForms {
			if progress != nil {
				progress("crawling for forms")
			}
			for _, f := range webqa.Crawl(ctx, h.httpClient, req.BaseURL, 15, 2).Forms {
				af := activesec.Form{Method: f.Method, Action: f.Action}
				for _, fld := range f.Fields {
					af.Fields = append(af.Fields, activesec.FormField{Name: fld.Name, Type: fld.Type, Value: fld.Value})
				}
				opt.Forms = append(opt.Forms, af)
			}
		}
		if progress != nil {
			progress("probing")
		}
		res := activesec.Run(ctx, h.httpClient, req.BaseURL, opt)
		var fail, review []string
		for _, f := range res.Findings {
			line := fmt.Sprintf("active %s: %s (%s)", f.ID, f.Title, f.Endpoint)
			switch f.Severity {
			case activesec.High:
				fail = append(fail, line)
			case activesec.Medium:
				review = append(review, line)
			}
		}
		return gin.H{"result": res, "gate": combine(nil, req.Unreviewed != "warn", capList(fail, 20), capList(review, 20))}, nil
	})
}

type loadRequest struct {
	URL         string            `json:"url" binding:"required"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	Stages      []int             `json:"stages"`
	StageSec    int               `json:"stage_sec"`
	MaxP95Ms    float64           `json:"max_p95_ms"`
	MaxErrRate  float64           `json:"max_error_rate"`
	Concurrency int               `json:"concurrency"`
	DurationSec int               `json:"duration_sec"`
	WindowSec   int               `json:"window_sec"`
}

func (h *Handler) loadConfig(c *gin.Context) (loadRequest, benchmark.Config, bool) {
	var req loadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return req, benchmark.Config{}, false
	}
	if err := h.checkTarget(req.URL); err != nil {
		badRequest(c, err)
		return req, benchmark.Config{}, false
	}
	return req, benchmark.Config{URL: req.URL, Method: req.Method, Headers: req.Headers, Body: req.Body}, true
}

func (h *Handler) jobStress(c *gin.Context) {
	req, cfg, ok := h.loadConfig(c)
	if !ok {
		return
	}
	stages := req.Stages
	if len(stages) == 0 {
		stages = []int{5, 10, 25, 50, 100}
	}
	if len(stages) > 12 {
		badRequest(c, fmt.Errorf("at most 12 stages"))
		return
	}
	maxErr := req.MaxErrRate
	if maxErr <= 0 {
		maxErr = 0.01
	}
	stageSec := clampInt(req.StageSec, 5, 60, 15)
	timeout := time.Duration(len(stages)*(stageSec+5)+30) * time.Second
	j := h.jobs.Start("stress", timeout, func(ctx context.Context, p func(string)) (any, error) {
		rep := benchmark.Stress(ctx, h.httpClient, cfg, stages, stageSec, req.MaxP95Ms, maxErr, p)
		var fail []string
		if rep.BreakingConcurrency > 0 {
			fail = append(fail, fmt.Sprintf("service degrades at %d concurrent users (max sustainable: %d)", rep.BreakingConcurrency, rep.MaxSustainable))
		}
		return gin.H{"report": rep, "gate": combine(nil, true, fail, nil)}, nil
	})
	c.JSON(http.StatusAccepted, gin.H{"job_id": j.ID, "status_url": "/api/v1/jobs/" + j.ID})
}

func (h *Handler) jobSoak(c *gin.Context) {
	req, cfg, ok := h.loadConfig(c)
	if !ok {
		return
	}
	cfg.Concurrency = clampInt(req.Concurrency, 1, benchmark.MaxConcurrency, 10)
	total := clampInt(req.DurationSec, 10, 3600, 600)
	window := clampInt(req.WindowSec, 5, 120, 30)
	j := h.jobs.Start("soak", time.Duration(total+120)*time.Second, func(ctx context.Context, p func(string)) (any, error) {
		rep := benchmark.Soak(ctx, h.httpClient, cfg, total, window, p)
		var fail []string
		if rep.Degraded {
			fail = append(fail, "soak: "+rep.Reason)
		}
		return gin.H{"report": rep, "gate": combine(nil, true, fail, nil)}, nil
	})
	c.JSON(http.StatusAccepted, gin.H{"job_id": j.ID, "status_url": "/api/v1/jobs/" + j.ID})
}

func clampInt(v, lo, hi, def int) int {
	if v == 0 {
		return def
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
