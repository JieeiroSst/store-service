package model

import "time"

type CouponUsage struct {
	ID             int64     `gorm:"column:id;primaryKey" json:"id"`
	CouponID       int64     `gorm:"column:coupon_id" json:"coupon_id"`
	UserID         int64     `gorm:"column:user_id" json:"user_id"`
	OrderID        int64     `gorm:"column:order_id" json:"order_id"`
	DiscountAmount float64   `gorm:"column:discount_amount" json:"discount_amount"`
	UsedAt         time.Time `gorm:"column:used_at" json:"used_at"`
}

func (CouponUsage) TableName() string { return "coupon_usage" }
