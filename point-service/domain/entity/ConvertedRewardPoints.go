package entity

import "time"

type ConvertedRewardPoint struct {
	RewConvertId          string    `gorm:"TYPE:BIGINT;PRIMARY_KEY;NOT NULL;COLUMN:rew_convert_id" json:"rew_convert_id"`
	RewConvertOrdDetailId string    `gorm:"NULL;COLUMN:rew_convert_ord_detail_id" json:"rew_convert_ord_detail_id"` //FK
	RewConvertDiscountId  string    `gorm:"NULL;COLUMN:rew_convert_discount_id" json:"rew_convert_discount_id"`     //FK
	RewConvertPoints      int       `gorm:"NOT NULL;COLUMN:rew_convert_points" json:"rew_convert_points"`
	RewConvertDate        int       `gorm:"NOT NULL;COLUMN:rew_convert_date" json:"rew_convert_date"`
	CreatedAt             time.Time `gorm:"NOT NULL;COLUMN:created_at" json:"created_at"`
	UpdatedAt             time.Time `gorm:"NOT NULL;COLUMN:updated_at" json:"updated_at"`
}

func (ConvertedRewardPoint) TableName() string {
	return "converted_reward_points"
}
