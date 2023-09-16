package entity

type RewardDiscount struct {
	RewardDiscountID int `gorm:"TYPE:BIGINT;PRIMARY_KEY;NOT NULL;COLUMN:reward_discount_id" json:"reward_discount_id"`
	TotalPoints      int `gorm:"NOT NULL;COLUMN:total_points" json:"total_points"`
	PointsPending    int `gorm:"NOT NULL;COLUMN:points_pending" json:"points_pending"`
	PointsActive     int `gorm:"NOT NULL;COLUMN:points_active" json:"points_active"`
	PointsExpired    int `gorm:"NOT NULL;COLUMN:points_expired" json:"points_expired"`
	PointsConverted  int `gorm:"NOT NULL;COLUMN:points_converted" json:"points_converted"`
	PointsCancelled  int `gorm:"NOT NULL;COLUMN:points_cancelled" json:"points_cancelled"`
	ActivateDate     int `gorm:"NOT NULL;COLUMN:activate_date" json:"activate_date"`
	ExpireDate       int `gorm:"NOT NULL;COLUMN:expire_date" json:"expire_date"`
}

func (RewardDiscount) TableName() string {
	return "reward_discounts"
}
