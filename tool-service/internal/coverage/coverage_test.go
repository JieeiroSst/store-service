package coverage

import "testing"

const doc = `Orders API
- The amount must be greater than zero.
- Requests without a token must be rejected with 401.
The service returns 201 when an order is created. Orders are pretty.
Refunds are allowed within 30 days of purchase.
Nội dung: Số tiền phải lớn hơn 0.`

func TestExtract(t *testing.T) {
	got := Extract(doc)
	if len(got) != 5 {
		t.Fatalf("want 5 testable statements, got %d: %q", len(got), got)
	}
	for _, s := range got {
		if s == "Orders are pretty" || s == "Orders API" {
			t.Errorf("non-normative sentence extracted: %q", s)
		}
	}
}

func TestCompute(t *testing.T) {
	reqs := Extract(doc)
	r := Compute(reqs, []string{
		"The amount must be greater than zero.",
		"requests without a token must be rejected with 401",
		"The service returns 201 when an order is created",
		"unrelated sentence that mentions nothing here",
	})
	if r.Covered != 3 || r.Total != 5 || r.Percent != 60 {
		t.Fatalf("%+v", r)
	}
	if len(r.Uncovered) != 2 {
		t.Fatalf("uncovered: %q", r.Uncovered)
	}
	if e := Compute(nil, nil); e.Total != 0 || e.Percent != 0 {
		t.Fatalf("%+v", e)
	}
}
