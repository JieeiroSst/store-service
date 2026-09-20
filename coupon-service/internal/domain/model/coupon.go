package model

import "time"

type CouponType string

const (
	CouponPercentage  CouponType = "percentage"
	CouponFixedAmount CouponType = "fixed_amount"
	CouponBuyXGetY    CouponType = "buy_x_get_y"
)

func (t CouponType) Valid() bool {
	switch t {
	case CouponPercentage, CouponFixedAmount, CouponBuyXGetY:
		return true
	default:
		return false
	}
}

type Coupon struct {
	ID                int64      `gorm:"column:id;primaryKey" json:"id"`
	Code              string     `gorm:"column:code" json:"code"`
	Type              CouponType `gorm:"column:type" json:"type"`
	DiscountValue     float64    `gorm:"column:discount_value" json:"discount_value"`
	MinimumPurchase   float64    `gorm:"column:minimum_purchase" json:"minimum_purchase"`
	MaxDiscountAmount *float64   `gorm:"column:max_discount_amount" json:"max_discount_amount,omitempty"`
	Description       string     `gorm:"column:description" json:"description,omitempty"`
	StartDate         time.Time  `gorm:"column:start_date" json:"start_date"`
	EndDate           time.Time  `gorm:"column:end_date" json:"end_date"`
	IsActive          bool       `gorm:"column:is_active" json:"is_active"`
	MaxUses           *int       `gorm:"column:max_uses" json:"max_uses,omitempty"`
	CurrentUses       int        `gorm:"column:current_uses" json:"current_uses"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (Coupon) TableName() string { return "coupons" }

// IsWithinWindow reports whether now falls within [StartDate, EndDate].
func (c Coupon) IsWithinWindow(now time.Time) bool {
	return !now.Before(c.StartDate) && !now.After(c.EndDate)
}

// HasUsesLeft reports whether the coupon can still be redeemed; a nil
// MaxUses means unlimited redemptions.
func (c Coupon) HasUsesLeft() bool {
	return c.MaxUses == nil || c.CurrentUses < *c.MaxUses
}

// CalculateDiscount applies the coupon's discount rule to a purchase amount,
// capped by MaxDiscountAmount (if set) and by the purchase amount itself.
func (c Coupon) CalculateDiscount(purchaseAmount float64) float64 {
	var discount float64
	switch c.Type {
	case CouponPercentage:
		discount = purchaseAmount * c.DiscountValue / 100
	case CouponFixedAmount, CouponBuyXGetY:
		// buy_x_get_y is approximated as a flat discount: this service has no
		// visibility into order line items, so the caller (e.g. basket/order
		// service) is expected to pre-compute the equivalent value of the
		// free item(s) into DiscountValue.
		discount = c.DiscountValue
	}
	if c.MaxDiscountAmount != nil && discount > *c.MaxDiscountAmount {
		discount = *c.MaxDiscountAmount
	}
	if discount > purchaseAmount {
		discount = purchaseAmount
	}
	if discount < 0 {
		discount = 0
	}
	return discount
}
