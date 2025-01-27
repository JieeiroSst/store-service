package dto

import "google.golang.org/protobuf/types/known/timestamppb"

type Coupon struct {
	Id                int64                  `json:"id,omitempty"`
	Code              string                 `json:"code,omitempty"`
	MinimumPurchase   float64                `json:"minimum_purchase,omitempty"`
	MaxDiscountAmount float64                `json:"max_discount_amount,omitempty"`
	Description       string                 `json:"description,omitempty"`
	StartDate         *timestamppb.Timestamp `json:"start_date,omitempty"`
	EndDate           *timestamppb.Timestamp `json:"end_date,omitempty"`
	IsActive          bool                   `json:"is_active,omitempty"`
	MaxUses           int32                  `json:"max_uses,omitempty"`
	CurrentUses       int32                  `json:"current_uses,omitempty"`
	CreatedAt         *timestamppb.Timestamp `json:"created_at,omitempty"`
	UpdatedAt         *timestamppb.Timestamp `json:"updated_at,omitempty"`
	DiscountValue     float64                `json:"discount_value,omitempty"`
	Type              string                 `json:"type,omitempty"`
}

type CouponRestriction struct {
	Id                 int64                  `json:"id,omitempty"`
	CouponId           int64                  `json:"coupon_id,omitempty"`
	RestrictionType    string                 `json:"restriction_type,omitempty"`
	RestrictedEntityId int64                  `json:"restricted_entity_id,omitempty"`
	IsExclude          bool                   `json:"is_exclude,omitempty"`
	CreatedAt          *timestamppb.Timestamp `json:"created_at,omitempty"`
}

type CouponUsage struct {
	Id             int64                  `json:"id,omitempty"`
	CouponId       int64                  `json:"coupon_id,omitempty"`
	UserId         int64                  `json:"user_id,omitempty"`
	OrderId        int64                  `json:"order_id,omitempty"`
	DiscountAmount float64                `json:"discount_amount,omitempty"`
	UsedAt         *timestamppb.Timestamp `json:"used_at,omitempty"`
}

type UserCoupon struct {
	Id         int64                  `json:"id,omitempty"`
	UserId     int64                  `json:"user_id,omitempty"`
	CouponId   int64                  `json:"coupon_id,omitempty"`
	IsUsed     bool                   `json:"is_used,omitempty"`
	AssignedAt *timestamppb.Timestamp `json:"assigned_at,omitempty"`
	UsedAt     *timestamppb.Timestamp `json:"used_at,omitempty"`
}

type CreateCouponRequest struct {
	Code              string                 `json:"code,omitempty"`
	MinimumPurchase   float64                `json:"minimum_purchase,omitempty"`
	MaxDiscountAmount float64                `json:"max_discount_amount,omitempty"`
	Description       string                 `json:"description,omitempty"`
	StartDate         *timestamppb.Timestamp `json:"start_date,omitempty"`
	EndDate           *timestamppb.Timestamp `json:"end_date,omitempty"`
	IsActive          bool                   `json:"is_active,omitempty"`
	MaxUses           int32                  `json:"max_uses,omitempty"`
}

type GetCouponRequest struct {
	Id int64 `json:"id,omitempty"`
}

type UpdateCouponRequest struct {
	Id                int64                  `json:"id,omitempty"`
	Code              string                 `json:"code,omitempty"`
	MinimumPurchase   float64                `json:"minimum_purchase,omitempty"`
	MaxDiscountAmount float64                `json:"max_discount_amount,omitempty"`
	Description       string                 `json:"description,omitempty"`
	StartDate         *timestamppb.Timestamp `json:"start_date,omitempty"`
	EndDate           *timestamppb.Timestamp `json:"end_date,omitempty"`
	IsActive          bool                   `json:"is_active,omitempty"`
	MaxUses           int32                  `json:"max_uses,omitempty"`
	DiscountValue     float64                `json:"discount_value,omitempty"`
	Type              string                 `json:"type,omitempty"`
}

type DeleteCouponRequest struct {
	Id int64 `json:"id,omitempty"`
}

type DeleteCouponRestrictionResponse struct {
	Success bool `json:"success,omitempty"`
}

type DeleteCouponResponse struct {
	Success bool `json:"success,omitempty"`
}

type ListCouponsRequest struct {
}

type ListCouponsResponse struct {
	Coupons []*Coupon `json:"coupons,omitempty"`
}

type CreateCouponRestrictionRequest struct {
	CouponId           int64  `json:"coupon_id,omitempty"`
	RestrictionType    string `json:"restriction_type,omitempty"`
	RestrictedEntityId int64  `json:"restricted_entity_id,omitempty"`
	IsExclude          bool   `json:"is_exclude,omitempty"`
}

type GetCouponRestrictionRequest struct {
	CouponId int64 `json:"coupon_id,omitempty"`
	Id       int64 `json:"id,omitempty"`
}

type UpdateCouponRestrictionRequest struct {
	CouponId           int64  `json:"coupon_id,omitempty"`
	Id                 int64  `json:"id,omitempty"`
	RestrictionType    string `json:"restriction_type,omitempty"`
	RestrictedEntityId int64  `json:"restricted_entity_id,omitempty"`
	IsExclude          bool   `json:"is_exclude,omitempty"`
}

type DeleteCouponRestrictionRequest struct {
	CouponId int64 `json:"coupon_id,omitempty"`
	Id       int64 `json:"id,omitempty"`
}

type ListCouponRestrictionsRequest struct {
	CouponId int64 `json:"coupon_id,omitempty"`
}

type ListCouponRestrictionsResponse struct {
	Restrictions []*CouponRestriction `json:"restrictions,omitempty"`
}

type CreateUserCouponRequest struct {
	UserId   int64 `json:"user_id,omitempty"`
	CouponId int64 `json:"coupon_id,omitempty"`
}

type GetUserCouponRequest struct {
	UserId int64 `json:"user_id,omitempty"`
	Id     int64 `json:"id,omitempty"`
}

type ListUserCouponsRequest struct {
	UserId int64 `json:"user_id,omitempty"`
}

type ListUserCouponsResponse struct {
	UserCoupons []*UserCoupon `json:"user_coupons,omitempty"`
}

type UseUserCouponRequest struct {
	UserId int64 `json:"user_id,omitempty"`
	Id     int64 `json:"id,omitempty"`
}

type ListCouponUsagesRequest struct {
	CouponId int64 `json:"coupon_id,omitempty"`
}

type ListCouponUsagesResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesRequest struct {
	UserId   int64 `json:"user_id,omitempty"`
	CouponId int64 `json:"coupon_id,omitempty"`
}

type UnuseUserCouponRequest struct {
	UserId int64 `json:"user_id,omitempty"`
	Id     int64 `json:"id,omitempty"`
}

type ListUserCouponUsagesResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByOrderRequest struct {
	UserId  int64 `json:"user_id,omitempty"`
	OrderId int64 `json:"order_id,omitempty"`
}

type ListUserCouponUsagesByOrderResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByOrderAndCouponRequest struct {
	UserId   int64 `json:"user_id,omitempty"`
	OrderId  int64 `json:"order_id,omitempty"`
	CouponId int64 `json:"coupon_id,omitempty"`
}

type ListUserCouponUsagesByCouponRequest struct {
	UserId   int64 `json:"user_id,omitempty"`
	CouponId int64 `json:"coupon_id,omitempty"`
}

type ListUserCouponUsagesByCouponResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByOrderAndCouponResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByOrderAndCouponAndUserRequest struct {
	UserId   int64 `json:"user_id,omitempty"`
	OrderId  int64 `json:"order_id,omitempty"`
	CouponId int64 `json:"coupon_id,omitempty"`
}

type ListUserCouponUsagesByOrderAndUserAndCouponRequest struct {
	UserId   int64 `json:"user_id,omitempty"`
	OrderId  int64 `json:"order_id,omitempty"`
	CouponId int64 `json:"coupon_id,omitempty"`
}

type ListUserCouponUsagesByOrderAndUserAndCouponResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByOrderAndCouponAndUserResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByOrderAndUserRequest struct {
	UserId  int64 `json:"user_id,omitempty"`
	OrderId int64 `json:"order_id,omitempty"`
}

type ListUserCouponUsagesByOrderAndUserResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByUserRequest struct {
	UserId int64 `json:"user_id,omitempty"`
}

type ListUserCouponUsagesByCouponAndUserRequest struct {
	UserId   int64 `json:"user_id,omitempty"`
	CouponId int64 `json:"coupon_id,omitempty"`
}

type ListUserCouponUsagesByCouponAndUserResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}

type ListUserCouponUsagesByUserResponse struct {
	Usages []*CouponUsage `json:"usages,omitempty"`
}
