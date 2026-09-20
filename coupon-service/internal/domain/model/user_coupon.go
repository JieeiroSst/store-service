package model

import "time"

type UserCoupon struct {
	ID         int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID     int64      `gorm:"column:user_id" json:"user_id"`
	CouponID   int64      `gorm:"column:coupon_id" json:"coupon_id"`
	IsUsed     bool       `gorm:"column:is_used" json:"is_used"`
	AssignedAt time.Time  `gorm:"column:assigned_at" json:"assigned_at"`
	UsedAt     *time.Time `gorm:"column:used_at" json:"used_at,omitempty"`
}

func (UserCoupon) TableName() string { return "user_coupons" }
