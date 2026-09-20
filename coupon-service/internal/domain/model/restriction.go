package model

import "time"

type RestrictionType string

const (
	RestrictionCategory  RestrictionType = "category"
	RestrictionProduct   RestrictionType = "product"
	RestrictionUserGroup RestrictionType = "user_group"
)

func (t RestrictionType) Valid() bool {
	switch t {
	case RestrictionCategory, RestrictionProduct, RestrictionUserGroup:
		return true
	default:
		return false
	}
}

type CouponRestriction struct {
	ID                 int64           `gorm:"column:id;primaryKey" json:"id"`
	CouponID           int64           `gorm:"column:coupon_id" json:"coupon_id"`
	RestrictionType    RestrictionType `gorm:"column:restriction_type" json:"restriction_type"`
	RestrictedEntityID int64           `gorm:"column:restricted_entity_id" json:"restricted_entity_id"`
	IsExclude          bool            `gorm:"column:is_exclude" json:"is_exclude"`
	CreatedAt          time.Time       `gorm:"column:created_at" json:"created_at"`
}

func (CouponRestriction) TableName() string { return "coupon_restrictions" }
