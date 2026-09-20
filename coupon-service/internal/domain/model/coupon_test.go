package model

import "testing"

func TestCouponCalculateDiscount(t *testing.T) {
	maxDiscount := 50000.0

	cases := []struct {
		name     string
		coupon   Coupon
		purchase float64
		want     float64
	}{
		{"percentage under cap", Coupon{Type: CouponPercentage, DiscountValue: 10}, 200000, 20000},
		{"percentage capped", Coupon{Type: CouponPercentage, DiscountValue: 50, MaxDiscountAmount: &maxDiscount}, 200000, 50000},
		{"fixed amount", Coupon{Type: CouponFixedAmount, DiscountValue: 30000}, 200000, 30000},
		{"fixed amount larger than purchase", Coupon{Type: CouponFixedAmount, DiscountValue: 30000}, 10000, 10000},
		{"buy x get y treated as flat value", Coupon{Type: CouponBuyXGetY, DiscountValue: 15000}, 200000, 15000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.coupon.CalculateDiscount(tc.purchase); got != tc.want {
				t.Errorf("CalculateDiscount(%v) = %v, want %v", tc.purchase, got, tc.want)
			}
		})
	}
}

func TestCouponHasUsesLeft(t *testing.T) {
	unlimited := Coupon{MaxUses: nil, CurrentUses: 1000}
	if !unlimited.HasUsesLeft() {
		t.Error("expected unlimited coupon to always have uses left")
	}

	max := 5
	exhausted := Coupon{MaxUses: &max, CurrentUses: 5}
	if exhausted.HasUsesLeft() {
		t.Error("expected coupon at max_uses to have no uses left")
	}

	available := Coupon{MaxUses: &max, CurrentUses: 4}
	if !available.HasUsesLeft() {
		t.Error("expected coupon under max_uses to have uses left")
	}
}

func TestCouponTypeValid(t *testing.T) {
	valid := []CouponType{CouponPercentage, CouponFixedAmount, CouponBuyXGetY}
	for _, ct := range valid {
		if !ct.Valid() {
			t.Errorf("expected %q to be valid", ct)
		}
	}
	if CouponType("bogus").Valid() {
		t.Error("expected unknown coupon type to be invalid")
	}
}
