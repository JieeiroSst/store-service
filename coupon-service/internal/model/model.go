package model

import (
	"time"
)

type Coupon struct {
	Id                int64     `json:"id,omitempty"`
	Code              string    `json:"code,omitempty"`
	Type              string    `json:"type,omitempty"`
	DiscountValue     float64   `json:"discount_value,omitempty"`
	MinimumPurchase   float64   `json:"minimum_purchase,omitempty"`
	MaxDiscountAmount float64   `json:"max_discount_amount,omitempty"`
	Description       string    `json:"description,omitempty"`
	StartDate         time.Time `json:"start_date,omitempty"`
	EndDate           time.Time `json:"end_date,omitempty"`
	IsActive          bool      `json:"is_active,omitempty"`
	MaxUses           int32     `json:"max_uses,omitempty"`
	CurrentUses       int32     `json:"current_uses,omitempty"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

type CouponRestriction struct {
	Id                 int64     `json:"id,omitempty"`
	CouponId           int64     `json:"coupon_id,omitempty"`
	RestrictionType    string    `json:"restriction_type,omitempty"`
	RestrictedEntityId int64     `json:"restricted_entity_id,omitempty"`
	IsExclude          bool      `json:"is_exclude,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
}

type CouponUsage struct {
	Id             int64     `json:"id,omitempty"`
	CouponId       int64     `json:"coupon_id,omitempty"`
	UserId         int64     `json:"user_id,omitempty"`
	OrderId        int64     `json:"order_id,omitempty"`
	DiscountAmount float64   `json:"discount_amount,omitempty"`
	UsedAt         time.Time `json:"used_at,omitempty"`
}

type UserCoupon struct {
	Id         int64     `json:"id,omitempty"`
	UserId     int64     `json:"user_id,omitempty"`
	CouponId   int64     `json:"coupon_id,omitempty"`
	IsUsed     bool      `json:"is_used,omitempty"`
	AssignedAt time.Time `json:"assigned_at,omitempty"`
	UsedAt     time.Time `json:"used_at,omitempty"`
}
