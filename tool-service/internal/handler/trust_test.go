package handler

import (
	"testing"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
)

const doc = "POST /orders creates an order.\nThe response status must be 201 Created.\nThe total is the sum of item prices."

func TestGrounding(t *testing.T) {
	for quote, want := range map[string]bool{
		"The response status must be 201 Created.": true,
		"the response  status MUST be 201 created": true,
		"The response status must be 200 OK.":      false,
		"":                                         false,
		"total":                                    false,
		"Refunds are returned within 5 business days.": false,
	} {
		if got := isGrounded(quote, doc); got != want {
			t.Errorf("isGrounded(%q) = %v, want %v", quote, got, want)
		}
	}
}

func TestMoneyDetection(t *testing.T) {
	yes := []apitest.TestCase{
		{Name: "pay", Path: "/orders", Body: map[string]any{"amount": 5}},
		{Name: "x", Path: "/wallet/transfer"},
		{Name: "tiền", Path: "/x"},
		{Name: "x", Path: "/x", ExpectJSON: map[string]any{"data.total": "calc:1+1"}},
	}
	for _, c := range yes {
		if !isMoneyRelated(c) {
			t.Errorf("%+v should be money related", c)
		}
	}
	if isMoneyRelated(apitest.TestCase{Name: "health", Path: "/health", ExpectStatus: 200}) {
		t.Error("health check is not money related")
	}
}

func TestMergeRunsVotes(t *testing.T) {
	a := apitest.TestCase{Name: "a", Path: "/a", ExpectStatus: 201}
	aRenamed := apitest.TestCase{Name: "same thing", Path: "/a", ExpectStatus: 201}
	b := apitest.TestCase{Name: "b", Path: "/b", ExpectStatus: 200}
	bDisagree := apitest.TestCase{Name: "b", Path: "/b", ExpectStatus: 404}
	cases, votes := mergeRuns([][]apitest.TestCase{{a, b}, {aRenamed}, {a, bDisagree, a}})
	if len(cases) != 3 {
		t.Fatalf("want 3 distinct assertions, got %d", len(cases))
	}
	if votes[caseKey(a)] != 3 {
		t.Errorf("a votes = %d, want 3 (duplicates within a run count once)", votes[caseKey(a)])
	}
	if votes[caseKey(b)] != 1 || votes[caseKey(bDisagree)] != 1 {
		t.Errorf("disagreeing expectations must not pool votes: %v", votes)
	}
}

func TestAssignTrust(t *testing.T) {
	ok := apitest.TestCase{Name: "created", Path: "/orders", Method: "POST", ExpectStatus: 201, SpecQuote: "The response status must be 201 Created."}
	invented := apitest.TestCase{Name: "invented", Path: "/orders", ExpectStatus: 200, SpecQuote: "Orders are free of charge."}
	money := apitest.TestCase{Name: "total", Path: "/orders", ExpectJSON: map[string]any{"total": "calc:19.99*3"}, SpecQuote: "The total is the sum of item prices."}
	badCalc := apitest.TestCase{Name: "badcalc", Path: "/x", ExpectJSON: map[string]any{"v": "calc:1+"}, SpecQuote: "POST /orders creates an order."}
	cases := []apitest.TestCase{ok, invented, money, badCalc}

	votes := map[string]int{caseKey(ok): 3, caseKey(invented): 3, caseKey(money): 3, caseKey(badCalc): 3}
	verified := map[int]bool{0: true, 1: true, 2: true, 3: true}
	tr := assignTrust(cases, doc, votes, 3, verified, map[string]bool{})
	want := []bool{true, false, false, false}
	for i, w := range want {
		if tr[i].Trusted != w {
			t.Errorf("%s trusted=%v want %v (%v)", tr[i].Name, tr[i].Trusted, w, tr[i].Reasons)
		}
	}
	if !tr[2].Critical {
		t.Error("money case must be critical")
	}

	tr = assignTrust(cases, doc, votes, 3, verified, map[string]bool{caseKey(money): true})
	if !tr[2].Trusted || !tr[2].Approved {
		t.Errorf("approved money case must be trusted: %+v", tr[2])
	}

	tr = assignTrust([]apitest.TestCase{ok}, doc, map[string]int{caseKey(ok): 1}, 3, verified, nil)
	if tr[0].Trusted {
		t.Error("1 of 3 agreeing runs must not be trusted")
	}
	tr = assignTrust([]apitest.TestCase{ok}, doc, map[string]int{caseKey(ok): 3}, 3, map[int]bool{0: false}, nil)
	if tr[0].Trusted {
		t.Error("verifier veto must remove trust")
	}
}
