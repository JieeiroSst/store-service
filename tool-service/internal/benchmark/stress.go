package benchmark

import (
	"context"
	"fmt"
	"net/http"
)

type StageResult struct {
	Concurrency int      `json:"concurrency"`
	Stats       Stats    `json:"stats"`
	Violations  []string `json:"violations,omitempty"`
}

type StressReport struct {
	Stages []StageResult `json:"stages"`

	BreakingConcurrency int     `json:"breaking_concurrency"`
	MaxSustainable      int     `json:"max_sustainable_concurrency"`
	PeakRPS             float64 `json:"peak_rps"`
}

func Stress(ctx context.Context, c *http.Client, cfg Config, stages []int, stageSec int, maxP95Ms, maxErrRate float64, progress func(string)) StressReport {
	rep := StressReport{}
	if stageSec <= 0 || stageSec > 60 {
		stageSec = 15
	}
	for i, conc := range stages {
		if ctx.Err() != nil {
			break
		}
		sc := cfg
		sc.Concurrency, sc.DurationSec, sc.Requests = conc, stageSec, 0
		if sc.Validate() != nil {
			break
		}
		if progress != nil {
			progress(fmt.Sprintf("stage %d/%d: %d concurrent", i+1, len(stages), conc))
		}
		st := Run(ctx, c, sc)
		sr := StageResult{Concurrency: conc, Stats: st}
		if st.TotalRequests > 0 {
			if e := float64(st.Failed) / float64(st.TotalRequests); maxErrRate > 0 && e > maxErrRate {
				sr.Violations = append(sr.Violations, fmt.Sprintf("error rate %.1f%% > %.1f%%", 100*e, 100*maxErrRate))
			}
		}
		if maxP95Ms > 0 && st.P95Ms > maxP95Ms {
			sr.Violations = append(sr.Violations, fmt.Sprintf("p95 %.1fms > %.1fms", st.P95Ms, maxP95Ms))
		}
		if st.RPS > rep.PeakRPS {
			rep.PeakRPS = st.RPS
		}
		rep.Stages = append(rep.Stages, sr)
		if len(sr.Violations) > 0 {
			rep.BreakingConcurrency = conc
			break
		}
		rep.MaxSustainable = conc
	}
	return rep
}

type Window struct {
	AtSec  int     `json:"at_sec"`
	RPS    float64 `json:"rps"`
	P95Ms  float64 `json:"p95_ms"`
	ErrPct float64 `json:"error_pct"`
}

type SoakReport struct {
	Windows  []Window `json:"windows"`
	Drift    Drift    `json:"drift"`
	Degraded bool     `json:"degraded"`
	Reason   string   `json:"reason,omitempty"`
}

type Drift struct {
	FirstQuarterP95Ms float64 `json:"first_quarter_p95_ms"`
	LastQuarterP95Ms  float64 `json:"last_quarter_p95_ms"`
	P95ChangePct      float64 `json:"p95_change_pct"`
	ErrPctChange      float64 `json:"error_pct_change"`
}

func Soak(ctx context.Context, c *http.Client, cfg Config, totalSec, windowSec int, progress func(string)) SoakReport {
	if totalSec < 10 || totalSec > 3600 {
		totalSec = 600
	}
	if windowSec < 5 || windowSec > 120 {
		windowSec = 30
	}
	return soak(ctx, c, cfg, totalSec, windowSec, progress)
}

func SoakWindows(ctx context.Context, c *http.Client, cfg Config, totalSec, windowSec int) SoakReport {
	return soak(ctx, c, cfg, totalSec, windowSec, nil)
}

func soak(ctx context.Context, c *http.Client, cfg Config, totalSec, windowSec int, progress func(string)) SoakReport {
	rep := SoakReport{}
	for at := 0; at < totalSec && ctx.Err() == nil; at += windowSec {
		wsec := windowSec
		if at+wsec > totalSec {
			wsec = totalSec - at
		}
		wc := cfg
		wc.DurationSec, wc.Requests = wsec, 0
		if wc.Validate() != nil {
			break
		}
		if progress != nil {
			progress(fmt.Sprintf("soak %ds/%ds", at, totalSec))
		}
		st := Run(ctx, c, wc)
		w := Window{AtSec: at, RPS: st.RPS, P95Ms: st.P95Ms}
		if st.TotalRequests > 0 {
			w.ErrPct = round(100 * float64(st.Failed) / float64(st.TotalRequests))
		}
		rep.Windows = append(rep.Windows, w)
	}
	analyzeDrift(&rep)
	return rep
}

func analyzeDrift(rep *SoakReport) {
	n := len(rep.Windows)
	if n < 4 {
		return
	}
	q := n / 4
	avg := func(ws []Window, f func(Window) float64) float64 {
		var s float64
		for _, w := range ws {
			s += f(w)
		}
		return s / float64(len(ws))
	}
	first, last := rep.Windows[:q], rep.Windows[n-q:]
	d := Drift{
		FirstQuarterP95Ms: round(avg(first, func(w Window) float64 { return w.P95Ms })),
		LastQuarterP95Ms:  round(avg(last, func(w Window) float64 { return w.P95Ms })),
		ErrPctChange:      round(avg(last, func(w Window) float64 { return w.ErrPct }) - avg(first, func(w Window) float64 { return w.ErrPct })),
	}
	if d.FirstQuarterP95Ms > 0 {
		d.P95ChangePct = round(100 * (d.LastQuarterP95Ms - d.FirstQuarterP95Ms) / d.FirstQuarterP95Ms)
	}
	rep.Drift = d
	switch {
	case d.P95ChangePct > 30 && d.LastQuarterP95Ms-d.FirstQuarterP95Ms > 20:
		rep.Degraded = true
		rep.Reason = fmt.Sprintf("p95 latency grew %.0f%% over the run (%.1fms -> %.1fms)", d.P95ChangePct, d.FirstQuarterP95Ms, d.LastQuarterP95Ms)
	case d.ErrPctChange > 1:
		rep.Degraded = true
		rep.Reason = fmt.Sprintf("error rate rose by %.1f points over the run", d.ErrPctChange)
	}
}
