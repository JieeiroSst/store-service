package domain

import (
	"encoding/json"
	"testing"
)

func TestDocumentQueryNormalize(t *testing.T) {
	q := DocumentQuery{
		Keyword: "  phone ",
		Indices: []string{"a, b", "", "book-store.*"},
		Filters: map[string][]string{"status": {"paid, pending"}, "empty": {""}},
		Ranges:  map[string]Range{"price": {Gte: "10"}, "none": {}},
		Sort:    ParseSort("price:desc, name"),
		Facets:  []string{"status,_index"},
	}
	if err := q.Normalize(ModeSearch); err != nil {
		t.Fatal(err)
	}
	if q.Keyword != "phone" || q.Size != DefaultPageSize || len(q.Indices) != 3 {
		t.Fatalf("unexpected %+v", q)
	}
	if len(q.Filters) != 1 || len(q.Filters["status"]) != 2 || len(q.Ranges) != 1 {
		t.Fatalf("filters %+v ranges %+v", q.Filters, q.Ranges)
	}
	if len(q.Sort) != 2 || !q.Sort[0].Desc || q.Sort[1].Desc || len(q.Facets) != 2 {
		t.Fatalf("sort %+v facets %+v", q.Sort, q.Facets)
	}

	bad := []DocumentQuery{
		{Keyword: "  "},
		{Keyword: "x", Size: MaxPageSize + 1},
		{Keyword: "x", Filters: map[string][]string{"_id": {"1"}}},
		{Keyword: "x", Sort: ParseSort("a,b,c,d")},
		{Keyword: "x", Facets: []string{"a b"}},
		{Keyword: "x", Indices: []string{".security"}},
		{Keyword: "x", Indices: []string{"-accounts"}},
		{Keyword: "x", Indices: []string{"Accounts"}},
	}
	for _, b := range bad {
		if err := b.Normalize(ModeSearch); !IsInvalid(err) {
			t.Errorf("expected invalid error for %+v, got %v", b, err)
		}
	}
	if err := (&DocumentQuery{}).Normalize(ModeList); err != nil {
		t.Errorf("list without keyword: %v", err)
	}
}

func TestDocumentRefValidate(t *testing.T) {
	if err := (DocumentRef{Index: "products", ID: "1"}).Validate(); err != nil {
		t.Error(err)
	}
	for _, ref := range []DocumentRef{{Index: "prod*", ID: "1"}, {Index: "products"}, {Index: "_all", ID: "1"}} {
		if err := ref.Validate(); !IsInvalid(err) {
			t.Errorf("%+v should be invalid", ref)
		}
	}
}

func TestSensitivePolicy(t *testing.T) {
	p, err := NewSensitivePolicy("")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"password", "password_hash", "refresh_token", "apiKey", "user.secret", "otp", "pin_code", "salt"} {
		if !p.IsSensitive(f) {
			t.Errorf("%s should be sensitive", f)
		}
	}
	for _, f := range []string{"name", "hotpot", "spin", "description", "shipping_note"} {
		if p.IsSensitive(f) {
			t.Errorf("%s should not be sensitive", f)
		}
	}

	doc := p.Redact(Document{
		Source: map[string]any{
			"name":     "a",
			"password": "x",
			"profile":  map[string]any{"api_key": "k", "city": "hn"},
			"cards":    []any{map[string]any{"card_number": "4111", "brand": "visa"}},
		},
		Highlight: map[string][]string{"password": {"x"}},
	})
	out, _ := json.Marshal(doc.Source)
	if string(out) != `{"cards":[{"brand":"visa"}],"name":"a","profile":{"city":"hn"}}` || doc.Highlight != nil {
		t.Errorf("redacted = %s highlight %v", out, doc.Highlight)
	}
	if err := p.CheckFields([]string{"name", "api_token"}); !IsInvalid(err) {
		t.Errorf("api_token should be rejected, got %v", err)
	}
	if _, err := NewSensitivePolicy("("); err == nil {
		t.Error("bad pattern should fail")
	}
}
