package entity

type ConvertedRewardPoints struct {
	RewConvertId          int `gorm:"TYPE:BIGINT;PRIMARY_KEY;NOT NULL;COLUMN:rew_convert_id" json:"rew_convert_id"`
	RewConvertOrdDetailId int `gorm:"NULL;COLUMN:rew_convert_ord_detail_id" json:"rew_convert_ord_detail_id"` //FK
	RewConvertDiscountId  int `gorm:"NULL;COLUMN:rew_convert_discount_id" json:"rew_convert_discount_id"`     //FK
	RewConvertPoints      int `gorm:"NOT NULL;COLUMN:rew_convert_points" json:"rew_convert_points"`
	RewConvertDate        int `gorm:"NOT NULL;COLUMN:rew_convert_date" json:"rew_convert_date"`
}
