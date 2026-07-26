package model

import "time"

type UserDevice struct {
	ID          uint      `json:"id" gorm:"column:id;primaryKey"`
	UserID      uint      `json:"user_id" gorm:"column:user_id"`
	DeviceToken string    `json:"device_token" gorm:"column:device_token"`
	DeviceType  string    `json:"device_type" gorm:"column:device_type"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active"`
	LastUsedAt  time.Time `json:"last_used_at" gorm:"column:last_used_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
}

func (UserDevice) TableName() string {
	return "notification_user_device"
}

type Notification struct {
	ID         uint      `json:"id" gorm:"column:id;primaryKey"`
	UserID     uint      `json:"user_id" gorm:"column:user_id"`
	Recipient  string    `json:"recipient" gorm:"column:recipient"`
	Title      string    `json:"title" gorm:"column:title"`
	Message    string    `json:"message" gorm:"column:message"`
	Type       string    `json:"type" gorm:"column:type"`
	Status     string    `json:"status" gorm:"column:status"`
	Priority   int       `json:"priority" gorm:"column:priority"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	SentAt     time.Time `json:"sent_at" gorm:"column:sent_at"`
	RetryCount int       `json:"retry_count" gorm:"column:retry_count"`
}

func (Notification) TableName() string {
	return "notification_notification"
}
