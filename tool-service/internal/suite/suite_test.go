package suite

import (
	"testing"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
)

func TestHashIgnoresNameAndQuote(t *testing.T) {
	a := apitest.TestCase{Name: "a", Method: "post", Path: "/x", ExpectStatus: 200, SpecQuote: "q1", ExpectJSON: map[string]any{"a": 1, "b": 2}}
	b := apitest.TestCase{Name: "b", Method: "POST", Path: "/x", ExpectStatus: 200, SpecQuote: "q2", ExpectJSON: map[string]any{"b": 2, "a": 1}}
	if Hash(a) != Hash(b) {
		t.Fatal("same assertions must hash equal")
	}
	b.ExpectJSON["a"] = 2
	if Hash(a) == Hash(b) {
		t.Fatal("different expectation must hash differently")
	}
}

func TestApproveDedupAndPrecision(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	c := apitest.TestCase{Name: "n", Path: "/x", ExpectStatus: 200}
	su, err := st.Approve("pay", []apitest.TestCase{c, c}, true, 1)
	if err != nil || len(su.Cases) != 1 || su.AIAccept != 1 || su.AIReject != 1 {
		t.Fatalf("got %+v %v", su, err)
	}
	if _, err := st.Approve("../evil", nil, false, 0); err == nil {
		t.Fatal("path traversal in suite name must be rejected")
	}
	got, _ := st.Get("pay")
	if len(got.Cases) != 1 {
		t.Fatal("suite not persisted")
	}
}

func TestPrecisionNeedsEvidenceFor99(t *testing.T) {

	if p := (&Suite{AIAccept: 50}).Precision(); p.Meets99 || p.Rate != 1 {
		t.Fatalf("50/50 must not meet 99%%: %+v", p)
	}
	if p := (&Suite{AIAccept: 500}).Precision(); !p.Meets99 {
		t.Fatalf("500/500 should meet 99%%: %+v", p)
	}
	if p := (&Suite{AIAccept: 490, AIReject: 10}).Precision(); p.Meets99 {
		t.Fatalf("98%% observed must not meet 99%%: %+v", p)
	}
}
