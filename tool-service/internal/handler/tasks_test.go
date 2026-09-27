package handler

import (
	"testing"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
)

func res(passed ...bool) *apitest.Report {
	r := &apitest.Report{Total: len(passed)}
	for _, p := range passed {
		r.Results = append(r.Results, apitest.Result{Passed: p})
		if p {
			r.Passed++
		} else {
			r.Failed++
		}
	}
	return r
}

func TestEvaluateGate(t *testing.T) {
	trusted := caseTrust{Name: "t", Trusted: true}
	untrusted := caseTrust{Name: "u"}
	money := caseTrust{Name: "m", Critical: true}
	moneyOK := caseTrust{Name: "mo", Critical: true, Approved: true, Trusted: true}
	slow := perfResult{Target: perfTarget{Path: "/a"}, Failures: []string{"p95 too high"}}

	cases := []struct {
		name       string
		res        taskResult
		block      bool
		wantStatus string
		wantPassed bool
	}{
		{"all trusted pass", taskResult{Report: res(true, true), Trust: []caseTrust{trusted, moneyOK}}, true, "passed", true},
		{"trusted case fails", taskResult{Report: res(true, false), Trust: []caseTrust{trusted, trusted}}, true, "failed", false},
		{"approved money case fails", taskResult{Report: res(false), Trust: []caseTrust{moneyOK}}, true, "failed", false},
		{"untrusted failure is review, not a bug", taskResult{Report: res(true, false), Trust: []caseTrust{trusted, untrusted}}, true, "needs_review", false},
		{"untrusted failure warn-only", taskResult{Report: res(true, false), Trust: []caseTrust{trusted, untrusted}}, false, "needs_review", true},
		{"unapproved money case never passes silently", taskResult{Report: res(true, true), Trust: []caseTrust{trusted, money}}, true, "needs_review", false},
		{"nothing trusted", taskResult{Report: res(true), Trust: []caseTrust{untrusted}}, true, "needs_review", false},
		{"slow endpoint fails", taskResult{Report: res(true), Trust: []caseTrust{trusted}, Performance: []perfResult{slow}}, true, "failed", false},
		{"no cases", taskResult{Report: res()}, true, "failed", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := evaluateGate(tc.res, 1.0, tc.block)
			if g.Status != tc.wantStatus || g.Passed != tc.wantPassed {
				t.Fatalf("status=%s passed=%v reasons=%v; want %s/%v", g.Status, g.Passed, g.Reasons, tc.wantStatus, tc.wantPassed)
			}
		})
	}
}
