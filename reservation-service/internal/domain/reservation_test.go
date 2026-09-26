package domain

import (
	"testing"
	"time"
)

func TestAddonAmountAndSum(t *testing.T) {
	tests := []struct {
		price string
		unit  AddonUnit
		qty   int
		night int
		want  string
	}{
		{"1500000.00", UnitStay, 2, 3, "3000000.0000"}, // airport transfer x2, once
		{"250000", UnitNight, 1, 3, "750000.0000"},     // breakfast, every night
		{"99.99", UnitNight, 2, 3, "599.9400"},         // exact decimal arithmetic
		{"0", UnitStay, 5, 1, "0.0000"},
	}
	for _, tc := range tests {
		got, err := AddonAmount(tc.price, tc.unit, tc.qty, tc.night)
		if err != nil || got != tc.want {
			t.Errorf("AddonAmount(%s,%s,%d,%d) = %q, %v; want %q", tc.price, tc.unit, tc.qty, tc.night, got, err, tc.want)
		}
	}
	if _, err := AddonAmount("-1", UnitStay, 1, 1); err == nil {
		t.Error("negative price must be rejected")
	}
	if sum, err := SumAmounts("0.1000", "0.2000", "1.0000"); err != nil || sum != "1.3000" {
		t.Errorf("sum = %q, %v", sum, err)
	}
	if _, err := SumAmounts("x"); err == nil {
		t.Error("bad amount")
	}
}

func TestCanRefundWithoutFee(t *testing.T) {
	start := time.Date(2027, 3, 10, 0, 0, 0, 0, time.UTC)
	hours := func(n int) time.Time { return start.Add(-time.Duration(n) * time.Hour) }
	if !CanRefundWithoutFee(hours(49), start, 48) {
		t.Error("49h before check-in with a 48h window is free")
	}
	if CanRefundWithoutFee(hours(48), start, 48) || CanRefundWithoutFee(hours(2), start, 48) {
		t.Error("inside the window is not free")
	}
	if !CanRefundWithoutFee(hours(1), start, 0) {
		t.Error("a 0h window allows cancelling until check-in")
	}
}

func TestReservationHelpers(t *testing.T) {
	r := Reservation{Start: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2027, 1, 4, 0, 0, 0, 0, time.UTC)}
	if r.Nights() != 3 {
		t.Error("nights")
	}
	for s, want := range map[ReservationStatus]string{1: "pending", 2: "paid", 3: "canceled", 4: "rejected", 5: "refunded", 9: ""} {
		if s.Name() != want {
			t.Errorf("status %d = %q", s, s.Name())
		}
	}
	if (InventoryDay{Total: 10, Reserved: 4}).Left() != 6 {
		t.Error("left")
	}
}

func TestRankScore(t *testing.T) {
	if RankScore(5, 1, 1) >= RankScore(4.8, 50, 1) {
		t.Fatal("one review outranks fifty")
	}
	if RankScore(4.5, 20, 4) <= RankScore(4.5, 20, 1) {
		t.Fatal("tier boost")
	}
	if RankScore(3.0, 100, 4) >= RankScore(4.6, 100, 1) {
		t.Fatal("boost overturned rating")
	}
	if BayesianRating(0, 0) != 3.5 || ManagerPoints(5) != 200 || ManagerPoints(4) != 100 || ManagerPoints(3) != 0 {
		t.Fatal("prior / points")
	}
}

func TestCancellationPolicy(t *testing.T) {
	policy := []CancellationTier{{HoursBefore: 72, FeePercent: 0}, {HoursBefore: 24, FeePercent: 50}}
	for hours, want := range map[float64]int{200: 0, 72: 0, 71.9: 50, 24: 50, 23.9: 100, 0: 100, -5: 100} {
		if got := FeePercentFor(policy, hours); got != want {
			t.Errorf("%.1fh before check-in: fee %d%%, want %d%%", hours, got, want)
		}
	}
	// The simple form (a free window only) is derived from free_cancel_hours.
	h := Hotel{FreeCancelHours: 48}
	if FeePercentFor(h.EffectivePolicy(), 60) != 0 || FeePercentFor(h.EffectivePolicy(), 47) != 100 {
		t.Error("default policy: free until 48h before, then non-refundable")
	}
	h.CancellationTiers = policy
	if len(h.EffectivePolicy()) != 2 {
		t.Error("hotel tiers win")
	}
}

func TestValidatePolicy(t *testing.T) {
	ok := []CancellationTier{{72, 0}, {24, 50}, {0, 90}}
	if err := ValidatePolicy(ok); err != nil {
		t.Fatal(err)
	}
	for name, p := range map[string][]CancellationTier{
		"hours not decreasing": {{24, 0}, {48, 50}},
		"same hours":           {{24, 0}, {24, 50}},
		"fee going down":       {{72, 50}, {24, 10}},
		"fee over 100":         {{72, 101}},
		"negative hours":       {{-1, 0}},
		"too many tiers":       {{100, 0}, {90, 10}, {80, 20}, {70, 30}, {60, 40}, {50, 50}},
	} {
		if err := ValidatePolicy(p); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestFeeAmountAndMinorUnits(t *testing.T) {
	for _, tc := range []struct {
		total, cur string
		pct        int
		want       string
	}{
		{"5000000.0000", "VND", 50, "2500000.0000"},
		{"33000000.0000", "VND", 30, "9900000.0000"},
		{"100.0000", "USD", 15, "15.0000"},
		{"99.9900", "USD", 50, "50.0000"}, // 49.995 rounds half up to the cent
		{"100.0000", "VND", 0, "0.0000"},
	} {
		got, err := FeeAmount(tc.total, tc.pct, tc.cur)
		if err != nil || got != tc.want {
			t.Errorf("FeeAmount(%s,%d%%,%s) = %q, %v; want %q", tc.total, tc.pct, tc.cur, got, err, tc.want)
		}
	}
	if FromMinorUnits(150000, "VND") != "150000.0000" || FromMinorUnits(1999, "usd") != "19.9900" {
		t.Error("FromMinorUnits")
	}
}

func TestPromotion(t *testing.T) {
	pct := Promotion{PercentOff: 15, MinNights: 3, Active: true}
	if d, err := pct.DiscountFor("10000000.0000"); err != nil || d != "1500000.0000" {
		t.Errorf("15%%: %q %v", d, err)
	}
	fixed := Promotion{AmountOff: "2000000", Active: true, MinNights: 1}
	if d, _ := fixed.DiscountFor("10000000.0000"); d != "2000000.0000" {
		t.Errorf("fixed: %q", d)
	}
	if d, _ := fixed.DiscountFor("1500000.0000"); d != "1500000.0000" {
		t.Errorf("never more than the charge: %q", d)
	}
	from, to := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC)
	p := Promotion{Active: true, MinNights: 3, MaxUses: 2, UsedCount: 1, ValidFrom: &from, ValidTo: &to}
	mid := time.Date(2027, 3, 10, 0, 0, 0, 0, time.UTC)
	if r := p.UsableFor(mid, 3); r != "" {
		t.Errorf("should be usable: %s", r)
	}
	for name, got := range map[string]string{
		"too short":    p.UsableFor(mid, 2),
		"too early":    p.UsableFor(from.AddDate(0, 0, -1), 3),
		"too late":     p.UsableFor(to.AddDate(0, 0, 1), 3),
		"used up":      Promotion{Active: true, MinNights: 1, MaxUses: 1, UsedCount: 1}.UsableFor(mid, 1),
		"switched off": Promotion{MinNights: 1}.UsableFor(mid, 1),
	} {
		if got == "" {
			t.Errorf("%s: must be refused", name)
		}
	}
}
