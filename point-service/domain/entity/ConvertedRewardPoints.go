package entity

type ConvertedRewardPoint struct {
	RewConvertId          int    `gorm:"TYPE:BIGINT;PRIMARY_KEY;NOT NULL;COLUMN:rew_convert_id" json:"rew_convert_id"`
	RewConvertOrdDetailId int    `gorm:"NULL;COLUMN:rew_convert_ord_detail_id" json:"rew_convert_ord_detail_id"` //FK
	RewConvertDiscountId  int    `gorm:"NULL;COLUMN:rew_convert_discount_id" json:"rew_convert_discount_id"`     //FK
	RewConvertPoints      int    `gorm:"NOT NULL;COLUMN:rew_convert_points" json:"rew_convert_points"`
	RewConvertDate        int    `gorm:"NOT NULL;COLUMN:rew_convert_date" json:"rew_convert_date"`
	CreatedAt             string `gorm:"NOT NULL;COLUMN:created_at" json:"created_at"`
	UpdatedAt             string `gorm:"NOT NULL;COLUMN:updated_at" json:"updated_at"`
}

func (ConvertedRewardPoint) TableName() string {
	return "converted_reward_points"
}
