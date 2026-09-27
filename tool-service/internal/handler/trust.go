package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
	"github.com/JIeeiroSst/tool-service/internal/suite"
)

type caseTrust struct {
	Name     string   `json:"name"`
	Hash     string   `json:"hash"`
	Trusted  bool     `json:"trusted"`
	Approved bool     `json:"approved"`
	Critical bool     `json:"critical"`
	Votes    int      `json:"votes"`
	Reasons  []string `json:"reasons,omitempty"`
}

var moneyWords = []string{
	"amount", "price", "money", "balance", "payment", "pay", "total", "fee", "tax", "discount", "refund", "currency",
	"cost", "charge", "invoice", "wallet", "transfer", "deposit", "withdraw", "cents", "vnd", "usd", "eur", "credit", "debit",
	"tiền", "tien", "giá", "gia", "thanh toán", "thanh toan", "phí", "phi", "thuế", "thue", "số dư", "so du",
	"hoàn tiền", "hoan tien", "chuyển khoản", "chuyen khoan", "nạp", "rút",
}

func isMoneyRelated(c apitest.TestCase) bool {
	b, _ := json.Marshal(struct {
		N, P string
		B, J any
		C    []string
	}{c.Name, c.Path, c.Body, c.ExpectJSON, c.ExpectBodyContains})
	text := strings.ToLower(string(b))
	for _, w := range moneyWords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func normalizeText(s string) string {
	var b strings.Builder
	space := true
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			space = false
		case !space:
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

func isGrounded(quote, doc string) bool {
	q := normalizeText(quote)
	return len(q) >= 12 && strings.Contains(normalizeText(doc), q)
}

func caseKey(c apitest.TestCase) string { return suite.Hash(c) }

func mergeRuns(runs [][]apitest.TestCase) ([]apitest.TestCase, map[string]int) {
	votes := map[string]int{}
	var out []apitest.TestCase
	for _, run := range runs {
		seen := map[string]bool{}
		for _, c := range run {
			k := caseKey(c)
			if seen[k] {
				continue
			}
			seen[k] = true
			if votes[k] == 0 {
				out = append(out, c)
			}
			votes[k]++
		}
	}
	return out, votes
}

const verifyPrompt = `You are an independent auditor. For each numbered test case decide whether the expectation is strictly and
unambiguously implied by the quoted requirement. Any doubt (unstated rounding, currency, error code, field name, or a value not
derivable from the quote and the request) means NOT supported. Reply ONLY with JSON:
{"results":[{"i":0,"supported":true,"reason":"short"}]}`

func (h *Handler) verify(ctx context.Context, cases []apitest.TestCase) map[int]bool {
	type item struct {
		I          int    `json:"i"`
		Quote      string `json:"requirement"`
		Method     string `json:"method"`
		Path       string `json:"path"`
		Body       any    `json:"request_body,omitempty"`
		Status     int    `json:"expect_status"`
		ExpectJSON any    `json:"expect_json,omitempty"`
	}
	var items []item
	for i, c := range cases {
		items = append(items, item{i, c.SpecQuote, c.Method, c.Path, c.Body, c.ExpectStatus, c.ExpectJSON})
	}
	payload, _ := json.Marshal(items)
	var out struct {
		Results []struct {
			I         int  `json:"i"`
			Supported bool `json:"supported"`
		} `json:"results"`
	}
	ok := map[int]bool{}
	if err := h.ai.ChatJSONWith(ctx, h.cfg.BaseModel, verifyPrompt, string(payload), &out); err != nil {
		return ok
	}
	for _, r := range out.Results {
		ok[r.I] = r.Supported
	}
	return ok
}

func assignTrust(cases []apitest.TestCase, doc string, votes map[string]int, runs int, verified map[int]bool, approved map[string]bool) []caseTrust {
	out := make([]caseTrust, len(cases))
	for i, c := range cases {
		k := caseKey(c)
		t := caseTrust{Name: c.Name, Hash: k, Votes: votes[k], Critical: isMoneyRelated(c), Approved: approved[k]}
		if t.Approved {
			t.Trusted = true
			out[i] = t
			continue
		}
		if !isGrounded(c.SpecQuote, doc) {
			t.Reasons = append(t.Reasons, "spec_quote missing or not found verbatim in the task document")
		}
		if t.Votes*2 <= runs {
			t.Reasons = append(t.Reasons, fmt.Sprintf("only %d/%d independent generations agreed on this case", t.Votes, runs))
		}
		if !verified[i] {
			t.Reasons = append(t.Reasons, "independent verifier did not confirm the expectation from the quote")
		}
		if err := checkCalc(c); err != nil {
			t.Reasons = append(t.Reasons, "invalid calc expression: "+err.Error())
		}
		if t.Critical {
			t.Reasons = append(t.Reasons, "money-related: requires human approval into a golden suite")
		}
		t.Trusted = len(t.Reasons) == 0
		out[i] = t
	}
	return out
}

func checkCalc(c apitest.TestCase) error {
	for _, want := range c.ExpectJSON {
		if _, err := apitest.ResolveExpected(want); err != nil {
			return err
		}
	}
	return nil
}
