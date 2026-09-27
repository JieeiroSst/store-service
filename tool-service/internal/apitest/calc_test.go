package apitest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvalCalcExact(t *testing.T) {
	for expr, want := range map[string]string{
		"1+1":            "2",
		"19.99*3":        "59.97",
		"0.1+0.2":        "0.3",
		"(100-15)*1.1":   "93.5",
		"100/3*3":        "100",
		"-5+2":           "-3",
		"1000000.01*100": "100000001",
	} {
		got, err := EvalCalc(expr)
		if err != nil {
			t.Fatalf("%s: %v", expr, err)
		}
		if w, _ := ResolveExpected("calc:" + want); !equalValues(got, w) {
			t.Errorf("%s = %s, want %s", expr, showValue(got), want)
		}
	}
	for _, bad := range []string{"", "1+", "1/0", "abc", "(1+2", "1+2)", "1 2"} {
		if _, err := EvalCalc(bad); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
}

func TestExactMoneyComparison(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"total":59.97,"tax":"0.30","big":12345678901234567.25}`))
	}))
	defer srv.Close()

	run := func(exp map[string]any) Result {
		rep := Run(context.Background(), srv.Client(), srv.URL, []TestCase{{Name: "t", Path: "/", ExpectJSON: exp}})
		return rep.Results[0]
	}
	if r := run(map[string]any{"total": "calc:19.99*3", "tax": "calc:0.1*3", "big": "12345678901234567.25"}); !r.Passed {
		t.Fatalf("expected pass, got %v", r.Failures)
	}

	if r := run(map[string]any{"total": "calc:19.99*3+0.01"}); r.Passed {
		t.Fatal("one cent difference must fail")
	}
	if r := run(map[string]any{"big": "12345678901234567.26"}); r.Passed {
		t.Fatal("precision beyond float64 must be compared exactly")
	}
	if r := run(map[string]any{"total": "calc:1+"}); r.Passed {
		t.Fatal("invalid expression must fail, not pass")
	}
}
