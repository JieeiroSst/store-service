package model

import "time"

type Outbound struct {
	ID       int `gorm:"primary_key"`
	Date     time.Time
	BookID   int
	Quantity int
}
