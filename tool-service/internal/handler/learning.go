package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
	"github.com/JIeeiroSst/tool-service/internal/benchmark"
	"github.com/JIeeiroSst/tool-service/internal/learning"
	"github.com/JIeeiroSst/tool-service/internal/ollama"
)

type LearnConfig struct {
	BaseModel  string
	EmbedModel string
	AgentName  string
	FewShot    int

	ConsistencyRuns int
}

func (h *Handler) registerLearning(v1 *gin.RouterGroup) {
	v1.POST("/qc/run", h.qcRun)
	l := v1.Group("/learning")
	l.POST("/feedback", h.feedback)
	l.GET("/stats", h.stats)
	l.GET("/pending", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": stripEmbeddings(h.mem.Pending(20))})
	})
	l.GET("/history", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"history": h.mem.History(500)}) })
	l.POST("/reflect", h.reflectEndpoint)
	l.POST("/evolve", h.evolve)
	l.GET("/export", h.export)
}

func (h *Handler) recall(ctx context.Context, kind, query, base string) (string, []float32) {
	emb, err := h.ai.Embed(ctx, h.cfg.EmbedModel, truncate(query, 4000))
	if err != nil {
		log.Printf("embed unavailable, falling back to recency: %v", err)
	}
	var b strings.Builder
	b.WriteString(base)
	if pb := h.mem.Playbook(); pb != "" {
		b.WriteString("\n\nLessons learned from past QC reviews (follow strictly):\n" + pb)
	}
	for i, e := range h.mem.Similar(kind, emb, h.cfg.FewShot) {
		fmt.Fprintf(&b, "\n\nApproved example %d\nInput:\n%s\nOutput:\n%s", i+1, truncate(e.Input, 1500), e.Target())
	}
	return b.String(), emb
}

func (h *Handler) record(kind, input string, output any, emb []float32, signals map[string]any) string {
	raw, _ := json.Marshal(output)
	e := &learning.Experience{Kind: kind, Input: input, Output: raw, Embedding: emb, Signals: signals, Model: h.ai.Model()}
	if err := h.mem.Add(e); err != nil {
		log.Printf("record experience: %v", err)
		return ""
	}
	go h.autoReview(e.ID, kind, input, raw, signals)
	return e.ID
}

func reportSignals(r apitest.Report) map[string]any {
	unexec := 0
	for _, res := range r.Results {
		if res.Status == 0 {
			unexec++
		}
	}
	return map[string]any{"total": r.Total, "passed": r.Passed, "failed": r.Failed, "unexecutable": unexec}
}

type qcRequest struct {
	Spec      string `json:"spec" binding:"required"`
	BaseURL   string `json:"base_url" binding:"required"`
	Count     int    `json:"count"`
	Benchmark bool   `json:"benchmark"`
}

const qcVerdictPrompt = `You are the QC lead signing off a release. From the test report (and benchmark if present) reply ONLY with JSON:
{"verdict":"pass|fail","confidence":0.0,"bugs":[{"title":"","evidence":"","severity":"low|medium|high|critical"}],
"likely_test_errors":["cases whose expectation is probably wrong rather than the API"],"performance":"","recommendations":[""]}`

func (h *Handler) qcRun(c *gin.Context) {
	var req qcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.checkTarget(req.BaseURL); err != nil {
		badRequest(c, err)
		return
	}
	if req.Count <= 0 || req.Count > 30 {
		req.Count = 12
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	var gen struct {
		Cases []apitest.TestCase `json:"cases"`
	}
	system, emb := h.recall(ctx, learning.KindQCRun, req.Spec, testCaseSystemPrompt)
	if err := h.ai.ChatJSON(ctx, system, fmt.Sprintf("Generate about %d test cases for:\n\n%s", req.Count, req.Spec), &gen); err != nil {
		aiError(c, err)
		return
	}
	report := apitest.Run(ctx, h.httpClient, req.BaseURL, gen.Cases)
	evidence := gin.H{"report": report}

	if req.Benchmark {
		for _, tc := range gen.Cases {

			if strings.EqualFold(tc.Method, "GET") || tc.Method == "" {
				for _, r := range report.Results {
					if r.Name == tc.Name && r.Passed {
						cfg := benchmark.Config{URL: strings.TrimRight(req.BaseURL, "/") + "/" + strings.TrimLeft(tc.Path, "/"), Headers: tc.Headers, Concurrency: 20, Requests: 300}
						if cfg.Validate() == nil {
							evidence["benchmark"] = benchmark.Run(ctx, h.httpClient, cfg)
						}
					}
				}
				if _, ok := evidence["benchmark"]; ok {
					break
				}
			}
		}
	}

	var verdict map[string]any
	ev, _ := json.Marshal(evidence)
	if err := h.ai.ChatJSON(ctx, qcVerdictPrompt, string(ev), &verdict); err != nil {
		verdict = map[string]any{"error": err.Error()}
	}

	result := gin.H{"cases": gen.Cases, "evidence": evidence, "verdict": verdict}
	id := h.record(learning.KindQCRun, req.Spec, result, emb, reportSignals(report))
	c.JSON(http.StatusOK, gin.H{"experience_id": id, "model": h.ai.Model(), "cases": gen.Cases, "evidence": evidence, "verdict": verdict,
		"next": "review the result and POST /api/v1/learning/feedback with experience_id so the agent learns"})
}

type feedbackRequest struct {
	ID        string          `json:"id" binding:"required"`
	Verdict   string          `json:"verdict" binding:"required"`
	Note      string          `json:"note"`
	Corrected json.RawMessage `json:"corrected"`
}

func (h *Handler) feedback(c *gin.Context) {
	var req feedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.mem.Feedback(req.ID, req.Verdict, req.Note, req.Corrected); err != nil {
		badRequest(c, err)
		return
	}
	h.mem.TakeSnapshot(h.ai.Model())
	c.JSON(http.StatusOK, h.mem.Stats())
}

const reflectPrompt = `You maintain the playbook of an AI QC engineer. Given the current playbook and recent human reviews
(verdict good/bad, reviewer notes, optional corrected output), rewrite the playbook as at most 15 short, concrete,
non-contradictory rules (checks to always do, mistakes to avoid, domain conventions). Output plain text bullets only.`

func (h *Handler) Reflect(ctx context.Context) error {
	recent := h.mem.WithFeedback(30)
	if len(recent) == 0 {
		return fmt.Errorf("no reviewed experiences yet")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Current playbook:\n%s\n\nRecent reviews:\n", h.mem.Playbook())
	for _, e := range recent {
		fmt.Fprintf(&b, "- kind=%s verdict=%s note=%q signals=%v input=%q\n", e.Kind, e.Verdict, e.Note, e.Signals, truncate(e.Input, 300))
		if len(e.Corrected) > 0 {
			fmt.Fprintf(&b, "  corrected=%s\n", truncate(string(e.Corrected), 500))
		}
	}
	text, err := h.ai.Chat(ctx, reflectPrompt, b.String(), false)
	if err != nil {
		return err
	}
	return h.mem.SetPlaybook(strings.TrimSpace(text))
}

func (h *Handler) reflectEndpoint(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()
	if err := h.Reflect(ctx); err != nil {
		aiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"playbook": h.mem.Playbook()})
}

const (
	minReviewsToCompare = 10
	rollbackMargin      = 0.05
)

func (h *Handler) AutoLearn(ctx context.Context, every time.Duration, minFeedback, evolveEvery int) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.mem.TakeSnapshot(h.ai.Model())
			h.guardRollback()

			if h.mem.FeedbackSinceReflect() >= minFeedback {
				rctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
				if err := h.Reflect(rctx); err != nil {
					log.Printf("auto reflect: %v", err)
				}
				cancel()
			}

			cur := h.ai.Model()
			onProbation := false
			if cur != h.cfg.BaseModel {
				r := h.mem.RateByModel(false)[cur]
				onProbation = r.Good+r.Bad < minReviewsToCompare
			}
			if evolveEvery > 0 && !onProbation && len(h.mem.Approved())-h.approvedAtEvolve >= evolveEvery {
				ectx, cancel := context.WithTimeout(ctx, 10*time.Minute)
				if name, n, err := h.Evolve(ectx); err != nil {
					log.Printf("auto evolve: %v", err)
				} else {
					log.Printf("evolved %s from %d examples", name, n)
				}
				cancel()
			}
		}
	}
}

func (h *Handler) Evolve(ctx context.Context) (string, int, error) {
	system := testCaseSystemPrompt
	if pb := h.mem.Playbook(); pb != "" {
		system += "\n\nLessons learned from past QC reviews (follow strictly):\n" + pb
	}
	var msgs []ollama.Message
	approved := h.mem.Approved()
	total := len(approved)
	if len(approved) > 8 {
		approved = approved[len(approved)-8:]
	}
	for _, e := range approved {
		msgs = append(msgs, ollama.Message{Role: "user", Content: truncate(e.Input, 1500)},
			ollama.Message{Role: "assistant", Content: string(e.Target())})
	}
	name := fmt.Sprintf("%s-v%s", h.cfg.AgentName, time.Now().UTC().Format("20060102150405"))
	if err := h.ai.Create(ctx, name, h.cfg.BaseModel, system, msgs); err != nil {
		return "", 0, err
	}
	h.ai.SetModel(name)
	_ = h.mem.SetActiveModel(name)
	h.approvedAtEvolve = total
	h.mem.TakeSnapshot(name)
	return name, len(approved), nil
}

func (h *Handler) evolve(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()
	name, n, err := h.Evolve(ctx)
	if err != nil {
		aiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"model": name, "base": h.cfg.BaseModel, "examples": n})
}

func (h *Handler) guardRollback() {
	cur := h.ai.Model()
	if cur == h.cfg.BaseModel {
		return
	}
	rates := h.mem.RateByModel(false)
	c, b := rates[cur], rates[h.cfg.BaseModel]
	if c.Good+c.Bad >= minReviewsToCompare && b.Good+b.Bad >= minReviewsToCompare && c.Rate < b.Rate-rollbackMargin {
		log.Printf("rollback: %s approval %.2f < base %.2f", cur, c.Rate, b.Rate)
		h.ai.SetModel(h.cfg.BaseModel)
		_ = h.mem.SetActiveModel("")
		h.mem.TakeSnapshot(h.cfg.BaseModel)
	}
}

func (h *Handler) export(c *gin.Context) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Content-Disposition", "attachment; filename=qc-finetune.jsonl")
	enc := json.NewEncoder(c.Writer)
	for _, e := range h.mem.Approved() {
		sys := testCaseSystemPrompt
		if e.Kind == learning.KindCode {
			sys = codeSystemPrompt
		}
		_ = enc.Encode(gin.H{"messages": []ollama.Message{
			{Role: "system", Content: sys},
			{Role: "user", Content: e.Input},
			{Role: "assistant", Content: string(e.Target())},
		}})
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

type statsResponse struct {
	learning.Stats
	ActiveModel string                        `json:"active_model"`
	BaseModel   string                        `json:"base_model"`
	ByModel     map[string]learning.ModelRate `json:"by_model"`
	HumanOnly   map[string]learning.ModelRate `json:"by_model_human_only"`
}

func (h *Handler) stats(c *gin.Context) {
	c.JSON(http.StatusOK, statsResponse{
		Stats: h.mem.Stats(), ActiveModel: h.ai.Model(), BaseModel: h.cfg.BaseModel,
		ByModel: h.mem.RateByModel(false), HumanOnly: h.mem.RateByModel(true),
	})
}

func stripEmbeddings(in []*learning.Experience) []learning.Experience {
	out := make([]learning.Experience, len(in))
	for i, e := range in {
		out[i] = *e
		out[i].Embedding = nil
	}
	return out
}

const judgePrompt = `You are a strict QC lead reviewing the work of a junior QC engineer (an AI). Score how good and trustworthy the work is.
Check: test cases match the described API and invent no endpoints/fields; expectations are consistent with the spec
(a failing case can be an API bug OR a wrong expectation - penalize wrong expectations); coverage of happy path, validation,
auth, boundaries and errors; for code reviews, every reported issue must really exist in the code, fixes must be correct,
and generated tests/benchmarks must be plausible and runnable.
Reply ONLY with JSON: {"score":0.0,"reason":"one sentence naming the main weakness or strength"} where score is 0..1.`

const (
	autoGoodScore = 0.8
	autoBadScore  = 0.3
)

func (h *Handler) autoReview(id, kind, input string, output json.RawMessage, signals map[string]any) {

	if total, _ := signals["total"].(int); total > 0 {
		if unexec, _ := signals["unexecutable"].(int); float64(unexec)/float64(total) > 0.3 {
			_ = h.mem.AutoReview(id, learning.VerdictBad, fmt.Sprintf("%d/%d cases could not be executed", unexec, total), 0.1)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	payload := fmt.Sprintf("Kind: %s\nMetrics: %v\n\nInput:\n%s\n\nWork produced:\n%s", kind, signals, truncate(input, 3000), truncate(string(output), 6000))
	var j struct {
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}

	if err := h.ai.ChatJSONWith(ctx, h.cfg.BaseModel, judgePrompt, payload, &j); err != nil {
		log.Printf("auto review %s: %v", id, err)
		return
	}
	verdict := ""
	switch {
	case j.Score >= autoGoodScore:
		verdict = learning.VerdictGood
	case j.Score <= autoBadScore:
		verdict = learning.VerdictBad
	}
	if err := h.mem.AutoReview(id, verdict, j.Reason, j.Score); err != nil {
		log.Printf("auto review %s: %v", id, err)
	}
}
