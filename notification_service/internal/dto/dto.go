package dto

import "time"

type UserDevice struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	DeviceToken string    `json:"device_token"`
	DeviceType  string    `json:"device_type"`
	IsActive    bool      `json:"is_active"`
	LastUsedAt  time.Time `json:"last_used_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Notification struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Priority   int       `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
	SentAt     time.Time `json:"sent_at"`
	RetryCount int       `json:"retry_count"`
}
