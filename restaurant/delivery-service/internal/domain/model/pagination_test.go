package model

import "testing"

func TestPaginationDefaults(t *testing.T) {
	var p Pagination
	if p.GetLimit() != 10 || p.GetPage() != 1 || p.GetOffset() != 0 || p.GetSort() != "id desc" {
		t.Fatalf("defaults = limit %d page %d offset %d sort %q", p.Limit, p.Page, p.GetOffset(), p.Sort)
	}
	p = Pagination{Limit: 20, Page: 3}
	if p.GetOffset() != 40 {
		t.Fatalf("offset = %d, want 40", p.GetOffset())
	}
	p = Pagination{Limit: -5, Page: -1}
	if p.GetLimit() != 10 || p.GetPage() != 1 {
		t.Fatalf("negative input not normalised: limit %d page %d", p.Limit, p.Page)
	}
}
