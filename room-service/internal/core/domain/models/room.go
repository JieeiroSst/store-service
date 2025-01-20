package models

import (
	"time"
)

type Room struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Name        string    `json:"name"`
	Password    string    `json:"password"`
	ActiveUsers int       `json:"activeUsers" gorm:"-"`
	CreatedAt   time.Time `json:"createdAt"`
}
