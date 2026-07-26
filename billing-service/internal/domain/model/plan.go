package model

type Plan struct {
	PlanID       int     `gorm:"column:plan_id;primaryKey;autoIncrement" json:"plan_id,omitempty"`
	Name         string  `gorm:"column:name" json:"name"`
	Description  string  `gorm:"column:description" json:"description"`
	Price        float64 `gorm:"column:price" json:"price"`
	BillingCycle string  `gorm:"column:billing_cycle" json:"billing_cycle"`
}

func (Plan) TableName() string { return "plans" }
