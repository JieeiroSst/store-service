package model

import "time"

type OrderTracking struct {
	ID        int64       `gorm:"column:id;primaryKey" json:"id"`
	OrderID   int64       `gorm:"column:order_id" json:"order_id"`
	Status    OrderStatus `gorm:"column:status" json:"status"`
	Latitude  *float64    `gorm:"column:latitude" json:"latitude,omitempty"`
	Longitude *float64    `gorm:"column:longitude" json:"longitude,omitempty"`
	Timestamp time.Time   `gorm:"column:timestamp" json:"timestamp"`
	Notes     string      `gorm:"column:notes" json:"notes,omitempty"`
}

func (OrderTracking) TableName() string { return "order_tracking" }
