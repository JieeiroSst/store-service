package usecase

import (
	"testing"
	"time"
)

func TestRentalFee(t *testing.T) {
	start := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		hours float64
		want  float64
	}{
		{"under an hour is billed as one hour", 0.2, 10},
		{"partial hours round up", 2.5, 30},
		{"remainder is capped at the daily rate", 20, 50},
		{"whole day", 24, 50},
		{"day plus hours", 27, 80},
	}
	for _, tt := range tests {
		end := start.Add(time.Duration(tt.hours * float64(time.Hour)))
		if got := rentalFee(start, end, 10, 50); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestPage(t *testing.T) {
	p, err := NewPage(0, "")
	if err != nil || p.Limit != defaultPageSize || p.Offset != 0 {
		t.Fatalf("defaults: %+v %v", p, err)
	}
	if p, _ = NewPage(1000, "40"); p.Limit != maxPageSize || p.Offset != 40 {
		t.Fatalf("clamp: %+v", p)
	}
	if _, err := NewPage(10, "-1"); err == nil {
		t.Fatal("negative token accepted")
	}
	if got := (Page{Offset: 0, Limit: 20}).NextToken(20, 45); got != "20" {
		t.Fatalf("next token = %q", got)
	}
	if got := (Page{Offset: 40, Limit: 20}).NextToken(5, 45); got != "" {
		t.Fatalf("last page token = %q", got)
	}
}

func TestLateFee(t *testing.T) {
	p := PricingPolicy{LateGrace: 15 * time.Minute, LateMultiplier: 1.5}.withDefaults()
	if got := p.lateFee(10*time.Minute, 10); got != 0 {
		t.Errorf("within grace: %v", got)
	}
	if got := p.lateFee(90*time.Minute, 10); got != 30 {
		t.Errorf("1.5h late = 2 started hours * 10 * 1.5: got %v", got)
	}
	if got := (PricingPolicy{}).withDefaults().lateFee(time.Hour, 10); got != 10 {
		t.Errorf("default policy: %v", got)
	}
}
