package handler

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
	"github.com/JIeeiroSst/tool-service/internal/benchmark"
	"github.com/JIeeiroSst/tool-service/internal/coverage"
	"github.com/JIeeiroSst/tool-service/internal/extract"
	"github.com/JIeeiroSst/tool-service/internal/learning"
	"github.com/JIeeiroSst/tool-service/internal/security"
	"github.com/JIeeiroSst/tool-service/internal/suite"
	"github.com/JIeeiroSst/tool-service/internal/webqa"
)

type perfTarget struct {
	Method      string  `json:"method"`
	Path        string  `json:"path"`
	P95Ms       float64 `json:"p95_ms"`
	MinRPS      float64 `json:"min_rps"`
	Concurrency int     `json:"concurrency"`
}

type taskPlan struct {
	Summary            string   `json:"summary"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Endpoints          []struct {
		Method      string `json:"method"`
		Path        string `json:"path"`
		Description string `json:"description"`
		Auth        string `json:"auth"`
	} `json:"endpoints"`
	CodeComponents []struct {
		Name             string   `json:"name"`
		Responsibilities string   `json:"responsibilities"`
		Risks            []string `json:"risks"`
	} `json:"code_components"`
	PerformanceTargets []perfTarget `json:"performance_targets"`
}

const taskCasePrompt = `You are a senior QA engineer writing API test cases that will gate a release for a payments-grade system.
Correctness matters more than coverage: a WRONG test case is far worse than a missing one.
Reply ONLY with JSON:
{"cases":[{"name":"","spec_quote":"","method":"GET","path":"/x","headers":{},"body":null,"expect_status":200,
"expect_body_contains":[],"expect_json":{"data.id":1,"data.total":"calc:19.99*3"},"max_latency_ms":500}]}
Hard rules:
1. spec_quote is REQUIRED: one sentence copied VERBATIM from the task document that states the behaviour under test.
   If you cannot quote the document for a case, do NOT write that case.
2. NEVER compute numbers yourself. Write any computed value (money, totals, tax, discounts) as "calc:<arithmetic expression>"
   using only numbers stated in the document or present in your request body, e.g. "calc:19.99*3". Only + - * / and parentheses.
3. Do not assume rounding, currency, status codes, error messages or field names the document does not state.
4. One behaviour per case; use concrete example values.`

const planPrompt = `You are a QC lead reading a task/requirement document. Decide what must be tested. Reply ONLY with JSON:
{"summary":"","acceptance_criteria":[""],
"endpoints":[{"method":"GET","path":"/orders/1","description":"","auth":"none|bearer|api-key|..."}],
"code_components":[{"name":"","responsibilities":"","risks":[""]}],
"performance_targets":[{"method":"GET","path":"/orders/1","p95_ms":300,"min_rps":100,"concurrency":20}]}
Use concrete example values in paths (e.g. /orders/1, never /orders/{id}). Take performance numbers from the document when
stated, otherwise propose sensible ones. Only include what the document actually describes.`

type perfResult struct {
	Target   perfTarget       `json:"target"`
	Stats    *benchmark.Stats `json:"stats,omitempty"`
	Passed   bool             `json:"passed"`
	Skipped  string           `json:"skipped,omitempty"`
	Failures []string         `json:"failures,omitempty"`
}

type gate struct {
	Passed bool `json:"passed"`

	Status      string   `json:"status"`
	Reasons     []string `json:"reasons,omitempty"`
	Trusted     int      `json:"trusted_cases"`
	Total       int      `json:"total_cases"`
	NeedsReview []string `json:"needs_review,omitempty"`
}

type taskResult struct {
	ExperienceID    string             `json:"experience_id,omitempty"`
	Model           string             `json:"model"`
	Source          string             `json:"source"`
	Plan            taskPlan           `json:"plan"`
	Cases           []apitest.TestCase `json:"cases"`
	Trust           []caseTrust        `json:"trust"`
	Suite           string             `json:"suite,omitempty"`
	Precision       *suite.Precision   `json:"suite_precision,omitempty"`
	Report          *apitest.Report    `json:"report,omitempty"`
	Performance     []perfResult       `json:"performance,omitempty"`
	Coverage        *coverage.Report   `json:"requirement_coverage,omitempty"`
	CoverageTrusted *coverage.Report   `json:"requirement_coverage_trusted,omitempty"`
	Security        []security.Finding `json:"security,omitempty"`
	Web             *webqa.Report      `json:"web,omitempty"`
	E2E             *e2eResult         `json:"e2e,omitempty"`
	Scope           []scopeItem        `json:"scope"`
	CodeAnalysis    map[string]any     `json:"code_analysis,omitempty"`
	Gate            *gate              `json:"gate,omitempty"`
}

func (h *Handler) analyzeTask(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, extract.MaxFileSize+(1<<20))

	var parts []string
	source := "text"
	if fh, err := c.FormFile("file"); err == nil {
		f, err := fh.Open()
		if err != nil {
			badRequest(c, err)
			return
		}
		data, err := io.ReadAll(io.LimitReader(f, extract.MaxFileSize+1))
		f.Close()
		if err != nil {
			badRequest(c, err)
			return
		}
		text, err := extract.Text(fh.Filename, data)
		if err != nil {
			badRequest(c, err)
			return
		}
		parts, source = append(parts, text), fh.Filename
	}
	if spec := strings.TrimSpace(c.PostForm("spec")); spec != "" {
		parts = append(parts, spec)
	}
	if len(parts) == 0 {
		badRequest(c, fmt.Errorf("provide a task document in form field \"file\" (.pdf/.docx/.txt/.md) or text in \"spec\""))
		return
	}
	text := strings.Join(parts, "\n\n")

	baseURL := strings.TrimSpace(c.PostForm("base_url"))
	if baseURL != "" {
		if err := h.checkTarget(baseURL); err != nil {
			badRequest(c, err)
			return
		}
	}
	minPass := formFloat(c, "min_pass_rate", 1.0)
	blockUnreviewed := c.DefaultPostForm("unreviewed", "block") != "warn"
	maxErr := formFloat(c, "max_error_rate", 0.01)
	format := c.DefaultPostForm("format", "json")
	if format != "json" && format != "junit" {
		badRequest(c, fmt.Errorf("format must be json or junit"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Minute)
	defer cancel()

	res := taskResult{Model: h.ai.Model(), Source: source}
	if err := h.ai.ChatJSON(ctx, planPrompt, "Task document:\n\n"+text, &res.Plan); err != nil {
		aiError(c, err)
		return
	}
	planJSON, _ := json.Marshal(res.Plan)

	system, emb := h.recall(ctx, learning.KindQCRun, text, taskCasePrompt)
	prompt := fmt.Sprintf("Write API test cases (about 15) that verify this task.\n\nTest plan:\n%s\n\nOriginal task document:\n%s",
		planJSON, truncate(text, 20000))

	runsWanted := clamp(formInt(c, "consistency_runs", h.cfg.ConsistencyRuns), 1, 5, 3)
	runs, err := h.generateRuns(ctx, system, prompt, runsWanted)
	if err != nil {
		aiError(c, err)
		return
	}
	cases, votes := mergeRuns(runs)

	reqs := coverage.Extract(text)
	if gaps := coverage.Compute(reqs, quotesOf(cases)).Uncovered; len(gaps) > 0 && c.PostForm("fill_gaps") != "false" {
		if gaps = firstN(gaps, 25); len(gaps) > 0 {
			gp := fmt.Sprintf("These requirements from the task document have NO test yet. Write test cases for them (same rules; skip any you cannot quote).\n\nUntested requirements:\n- %s\n\nTest plan:\n%s\n\nOriginal task document:\n%s",
				strings.Join(gaps, "\n- "), planJSON, truncate(text, 20000))
			if gapRuns, err := h.generateRuns(ctx, system, gp, runsWanted); err == nil {
				extra, extraVotes := mergeRuns(gapRuns)
				for _, ec := range extra {
					k := caseKey(ec)
					if _, dup := votes[k]; !dup {
						cases = append(cases, ec)
					}
					votes[k] += extraVotes[k]
				}
			}
		}
	}
	if len(cases) > 60 {
		cases = cases[:60]
	}

	suiteName := strings.TrimSpace(c.PostForm("suite"))
	approved := map[string]bool{}
	if suiteName != "" {
		if h.suites == nil || !suite.ValidName(suiteName) {
			badRequest(c, fmt.Errorf("invalid suite name"))
			return
		}
		su, err := h.suites.Get(suiteName)
		if err != nil {
			badRequest(c, err)
			return
		}
		res.Suite = suiteName
		p := su.Precision()
		res.Precision = &p
		have := map[string]bool{}
		for _, cs := range cases {
			have[caseKey(cs)] = true
		}
		for _, a := range su.Cases {
			approved[a.Hash] = true
			if !have[a.Hash] {
				cases = append(cases, a.Case)
				votes[a.Hash] = len(runs)
			}
		}
	}

	var pending []apitest.TestCase
	var pendingIdx []int
	for i, cs := range cases {
		if !approved[caseKey(cs)] {
			pending = append(pending, cs)
			pendingIdx = append(pendingIdx, i)
		}
	}
	verified := map[int]bool{}
	if len(pending) > 0 {
		for j, ok := range h.verify(ctx, pending) {
			if j >= 0 && j < len(pendingIdx) {
				verified[pendingIdx[j]] = ok
			}
		}
	}
	res.Cases = cases
	res.Trust = assignTrust(cases, text, votes, len(runs), verified, approved)
	cov := coverage.Compute(reqs, quotesOf(cases))
	covT := coverage.Compute(reqs, trustedQuotes(cases, res.Trust))
	res.Coverage, res.CoverageTrusted = &cov, &covT
	res.Scope = scopeMatrix()

	if c.PostForm("e2e") == "true" {
		res.E2E = h.generateE2E(ctx, planJSON, text)
	}

	signals := map[string]any{}
	if baseURL != "" {
		report := apitest.Run(ctx, h.httpClient, baseURL, res.Cases)
		res.Report = &report
		signals = reportSignals(report)
		if c.DefaultPostForm("benchmark", "true") != "false" {
			res.Performance = h.runPerformance(ctx, baseURL, res.Plan.PerformanceTargets, maxErr)
		}
		res.Gate = evaluateGate(res, minPass, blockUnreviewed)

		if c.DefaultPostForm("security", "true") != "false" {
			res.Security = security.Scan(ctx, h.httpClient, baseURL, security.Options{ProtectedPaths: protectedPaths(res.Plan)})
		}
		if c.DefaultPostForm("crawl", "true") != "false" {
			w := webqa.Crawl(ctx, h.httpClient, baseURL, clamp(formInt(c, "max_pages", 20), 1, 100, 20), 2)
			res.Web = &w
		}
		applyProductChecks(res.Gate, res.Security, res.Web, blockUnreviewed)
	}

	if code := strings.TrimSpace(c.PostForm("code")); code != "" {
		var out map[string]any
		csys, _ := h.recall(ctx, learning.KindCode, code, codeSystemPrompt)
		cp := fmt.Sprintf("Language: %s\nTask context (components and risks to focus on):\n%s\n\n```\n%s\n```",
			orDefault(c.PostForm("language"), "go"), planJSON, truncate(code, 20000))
		if err := h.ai.ChatJSON(ctx, csys, cp, &out); err != nil {
			out = map[string]any{"error": err.Error()}
		}
		res.CodeAnalysis = out
	}

	res.ExperienceID = h.record(learning.KindQCRun, text, res, emb, signals)

	status := http.StatusOK
	if res.Gate != nil && !res.Gate.Passed && c.PostForm("fail_http") == "true" {
		status = http.StatusUnprocessableEntity
	}
	if format == "junit" {
		c.Data(status, "application/xml", toJUnit(res))
		return
	}
	c.JSON(status, res)
}

func (h *Handler) runPerformance(ctx context.Context, baseURL string, targets []perfTarget, maxErr float64) []perfResult {
	if len(targets) > 5 {
		targets = targets[:5]
	}
	var out []perfResult
	for _, t := range targets {
		pr := perfResult{Target: t}
		method := strings.ToUpper(orDefault(t.Method, "GET"))
		switch {
		case method != http.MethodGet && method != http.MethodHead:
			pr.Skipped = "only read-only (GET/HEAD) endpoints are load-tested automatically"
		case strings.ContainsAny(t.Path, "{}:"):
			pr.Skipped = "path has unresolved parameters"
		default:
			cfg := benchmark.Config{
				URL: strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(t.Path, "/"), Method: method,
				Concurrency: clamp(t.Concurrency, 1, 50, 10), Requests: 300,
			}
			if err := cfg.Validate(); err != nil {
				pr.Skipped = err.Error()
				break
			}
			st := benchmark.Run(ctx, h.httpClient, cfg)
			pr.Stats = &st
			if st.TotalRequests > 0 && float64(st.Failed)/float64(st.TotalRequests) > maxErr {
				pr.Failures = append(pr.Failures, fmt.Sprintf("error rate %.1f%% exceeds %.1f%%",
					100*float64(st.Failed)/float64(st.TotalRequests), 100*maxErr))
			}
			if t.P95Ms > 0 && st.P95Ms > t.P95Ms {
				pr.Failures = append(pr.Failures, fmt.Sprintf("p95 %.1fms exceeds target %.1fms", st.P95Ms, t.P95Ms))
			}
			if t.MinRPS > 0 && st.RPS < t.MinRPS {
				pr.Failures = append(pr.Failures, fmt.Sprintf("throughput %.1f rps below target %.1f rps", st.RPS, t.MinRPS))
			}
			pr.Passed = len(pr.Failures) == 0
		}
		out = append(out, pr)
	}
	return out
}

func evaluateGate(res taskResult, minPass float64, blockUnreviewed bool) *gate {
	g := &gate{Passed: true, Status: "passed"}
	r := res.Report
	if r == nil || r.Total == 0 {
		g.Passed, g.Status = false, "failed"
		g.Reasons = append(g.Reasons, "no test cases were generated")
		return g
	}
	g.Total = r.Total
	var trustedPassed int
	var trustedFailed, untrustedFailed, unapprovedCritical []string
	for i, tr := range res.Trust {
		if i >= len(r.Results) {
			break
		}
		passed := r.Results[i].Passed
		switch {
		case tr.Trusted:
			g.Trusted++
			if passed {
				trustedPassed++
			} else {
				trustedFailed = append(trustedFailed, tr.Name)
			}
		case !passed:
			untrustedFailed = append(untrustedFailed, tr.Name)
		}
		if tr.Critical && !tr.Approved {
			unapprovedCritical = append(unapprovedCritical, tr.Name)
		}
	}

	failed := false
	if g.Trusted > 0 {
		if rate := float64(trustedPassed) / float64(g.Trusted); rate < minPass {
			failed = true
			g.Reasons = append(g.Reasons, fmt.Sprintf("trusted test pass rate %.0f%% below required %.0f%% (failed: %s)",
				100*rate, 100*minPass, strings.Join(trustedFailed, ", ")))
		}
	}
	for _, p := range res.Performance {
		if p.Skipped == "" && !p.Passed {
			failed = true
			g.Reasons = append(g.Reasons, fmt.Sprintf("performance %s %s: %s", orDefault(p.Target.Method, "GET"), p.Target.Path, strings.Join(p.Failures, "; ")))
		}
	}
	if failed {
		g.Passed, g.Status = false, "failed"
		return g
	}

	if g.Trusted == 0 {
		g.NeedsReview = append(g.NeedsReview, "no case could be trusted yet; approve reviewed cases into a golden suite")
	}
	if len(untrustedFailed) > 0 {
		g.NeedsReview = append(g.NeedsReview, "untrusted cases failed (may be test errors or real bugs): "+strings.Join(untrustedFailed, ", "))
	}
	if len(unapprovedCritical) > 0 {
		g.NeedsReview = append(g.NeedsReview, "money-related cases not yet approved by a human: "+strings.Join(unapprovedCritical, ", "))
	}
	if len(g.NeedsReview) > 0 {
		g.Status = "needs_review"
		g.Reasons = append(g.Reasons, g.NeedsReview...)
		g.Passed = !blockUnreviewed
	}
	return g
}

type junitSuites struct {
	XMLName xml.Name     `xml:"testsuites"`
	Suites  []junitSuite `xml:"testsuite"`
}
type junitSuite struct {
	Name     string      `xml:"name,attr"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Cases    []junitCase `xml:"testcase"`
}
type junitCase struct {
	Name    string   `xml:"name,attr"`
	Time    float64  `xml:"time,attr"`
	Failure *failure `xml:"failure,omitempty"`
	Skipped *skipped `xml:"skipped,omitempty"`
}
type skipped struct {
	Message string `xml:"message,attr,omitempty"`
}

type failure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

func toJUnit(res taskResult) []byte {
	api := junitSuite{Name: "api-tests"}
	if res.Report != nil {
		for i, r := range res.Report.Results {
			jc := junitCase{Name: r.Name, Time: float64(r.LatencyMs) / 1000}
			trusted := i >= len(res.Trust) || res.Trust[i].Trusted
			switch {
			case r.Passed:
			case trusted:
				jc.Failure = &failure{Message: strings.Join(r.Failures, "; "), Text: r.BodySample}
				api.Failures++
			default:
				jc.Skipped = &skipped{Message: "needs review (untrusted case): " + strings.Join(r.Failures, "; ")}
			}
			api.Cases = append(api.Cases, jc)
		}
		api.Tests = len(api.Cases)
	}
	perf := junitSuite{Name: "performance"}
	for _, p := range res.Performance {
		jc := junitCase{Name: fmt.Sprintf("%s %s", orDefault(p.Target.Method, "GET"), p.Target.Path)}
		switch {
		case p.Skipped != "":
			jc.Skipped = &skipped{Message: p.Skipped}
		case !p.Passed:
			jc.Failure = &failure{Message: strings.Join(p.Failures, "; ")}
			perf.Failures++
		}
		if p.Stats != nil {
			jc.Time = float64(p.Stats.ElapsedMs) / 1000
		}
		perf.Cases = append(perf.Cases, jc)
	}
	perf.Tests = len(perf.Cases)

	sec := junitSuite{Name: "security"}
	for _, f := range res.Security {
		if f.Severity != security.High && f.Severity != security.Medium {
			continue
		}
		jc := junitCase{Name: f.ID + " " + f.Title}
		if f.Severity == security.High {
			jc.Failure = &failure{Message: f.Detail, Text: f.Fix}
			sec.Failures++
		} else {
			jc.Skipped = &skipped{Message: "needs review: " + f.Detail}
		}
		sec.Cases = append(sec.Cases, jc)
	}
	sec.Tests = len(sec.Cases)

	web := junitSuite{Name: "web"}
	if res.Web != nil {
		for _, b := range res.Web.BrokenLinks {
			web.Cases = append(web.Cases, junitCase{Name: "broken link " + b.URL, Failure: &failure{Message: fmt.Sprintf("linked from %s, status %d %s", b.From, b.Status, b.Error)}})
			web.Failures++
		}
		for _, p := range res.Web.Pages {
			for _, is := range p.Issues {
				if is.Severity == webqa.Error {
					web.Cases = append(web.Cases, junitCase{Name: p.URL + " " + is.Rule, Failure: &failure{Message: is.Detail}})
					web.Failures++
				}
			}
		}
	}
	web.Tests = len(web.Cases)
	out, _ := xml.MarshalIndent(junitSuites{Suites: []junitSuite{api, perf, sec, web}}, "", "  ")
	return append([]byte(xml.Header), out...)
}

func formInt(c *gin.Context, key string, def int) int {
	if v, err := strconv.Atoi(c.PostForm(key)); err == nil {
		return v
	}
	return def
}

func formFloat(c *gin.Context, key string, def float64) float64 {
	if v, err := strconv.ParseFloat(c.PostForm(key), 64); err == nil && v >= 0 {
		return v
	}
	return def
}

func clamp(v, lo, hi, def int) int {
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
