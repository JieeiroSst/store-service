package model

import "time"

type Inventory struct {
	ID       int `gorm:"primary_key"`
	Date     time.Time
	BookID   int
	Quantity int
}
