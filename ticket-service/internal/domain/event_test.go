package domain

import (
	"testing"
	"time"
)

func TestPromotionDiscount(t *testing.T) {
	cases := []struct {
		p        Promotion
		subtotal int64
		want     int64
	}{
		{Promotion{Kind: PromoPercent, Value: 10}, 1_200_000, 120_000},
		{Promotion{Kind: PromoPercent, Value: 100}, 500, 500},
		{Promotion{Kind: PromoPercent, Value: 33}, 100, 33}, // rounds down: never gives away more than promised
		{Promotion{Kind: PromoFixed, Value: 50_000}, 1_200_000, 50_000},
		{Promotion{Kind: PromoFixed, Value: 50_000}, 20_000, 20_000}, // never below zero
		{Promotion{Kind: PromoFixed, Value: 50_000}, 0, 0},
	}
	for _, c := range cases {
		if got := c.p.DiscountFor(c.subtotal); got != c.want {
			t.Errorf("%+v on %d = %d, want %d", c.p, c.subtotal, got, c.want)
		}
	}
}

func TestPromotionUsableFor(t *testing.T) {
	now := time.Now()
	hour := time.Hour
	base := Promotion{Active: true, MinTickets: 2}
	if why := base.UsableFor(now, 2); why != "" {
		t.Fatalf("a fresh code: %q", why)
	}
	for name, p := range map[string]Promotion{
		"inactive":     {Active: false, MinTickets: 1},
		"not yet":      {Active: true, MinTickets: 1, ValidFrom: ptr(now.Add(hour))},
		"expired":      {Active: true, MinTickets: 1, ValidTo: ptr(now.Add(-hour))},
		"fully used":   {Active: true, MinTickets: 1, MaxUses: 3, Used: 3},
		"too few":      {Active: true, MinTickets: 4},
		"way over use": {Active: true, MinTickets: 1, MaxUses: 1, Used: 5},
	} {
		if why := p.UsableFor(now, 3); why == "" {
			t.Errorf("%s: code was usable", name)
		}
	}
	if why := (Promotion{Active: true, MinTickets: 1, MaxUses: 0, Used: 999}).UsableFor(now, 1); why != "" {
		t.Errorf("max_uses 0 means unlimited: %q", why)
	}
}

func ptr[T any](v T) *T { return &v }

func TestTicketTypeOnSale(t *testing.T) {
	now := time.Now()
	tt := TicketType{Active: true, Total: 10, Available: 4, Sold: 5}
	if tt.Held() != 1 {
		t.Fatalf("held = %d", tt.Held())
	}
	if !tt.OnSale(now) {
		t.Fatal("no window means always on sale")
	}
	tt.SaleStartsAt = ptr(now.Add(time.Hour))
	if tt.OnSale(now) {
		t.Fatal("not started")
	}
	tt.SaleStartsAt, tt.SaleEndsAt = ptr(now.Add(-2*time.Hour)), ptr(now.Add(-time.Hour))
	if tt.OnSale(now) {
		t.Fatal("ended")
	}
	tt.SaleEndsAt, tt.Active = nil, false
	if tt.OnSale(now) {
		t.Fatal("inactive")
	}
}

func TestEventRefundable(t *testing.T) {
	now := time.Now()
	e := Event{StartsAt: now.Add(72 * time.Hour), RefundCutoffHours: 48}
	if !e.Refundable(now) {
		t.Fatal("72h ahead with a 48h cutoff is refundable")
	}
	e.StartsAt = now.Add(24 * time.Hour)
	if e.Refundable(now) {
		t.Fatal("24h ahead with a 48h cutoff is not")
	}
	e.RefundCutoffHours = 0
	e.StartsAt = now.Add(1000 * time.Hour)
	if e.Refundable(now) {
		t.Fatal("cutoff 0 means never refundable")
	}
}

func TestOnSale(t *testing.T) {
	now := time.Now()
	e := Event{Status: EventPublished, EndsAt: now.Add(time.Hour)}
	if !e.OnSale(now) {
		t.Fatal("published and not over")
	}
	e.EndsAt = now.Add(-time.Minute)
	if e.OnSale(now) {
		t.Fatal("over")
	}
	e.EndsAt, e.Status = now.Add(time.Hour), EventCancelled
	if e.OnSale(now) {
		t.Fatal("cancelled")
	}
}

func TestSoldOutErrorMatchesSentinel(t *testing.T) {
	var err error = &SoldOutError{TicketTypeID: 5, Name: "VIP"}
	if !isErr(err, ErrSoldOut) || isErr(err, ErrConflict) {
		t.Fatal("SoldOutError must be ErrSoldOut and nothing else")
	}
}

func isErr(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		if is, ok := err.(interface{ Is(error) bool }); ok && is.Is(target) {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
