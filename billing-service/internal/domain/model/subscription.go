package model

type Subscription struct {
	SubscriptionID int    `gorm:"column:subscription_id;primaryKey;autoIncrement" json:"subscription_id,omitempty"`
	CustomerID     int    `gorm:"column:customer_id" json:"customer_id"`
	PlanID         int    `gorm:"column:plan_id" json:"plan_id"`
	StartDate      int64  `gorm:"column:start_date" json:"start_date"`
	EndDate        int64  `gorm:"column:end_date" json:"end_date,omitempty"`
	Status         string `gorm:"column:status" json:"status"`
}

func (Subscription) TableName() string { return "subscriptions" }
