package entity

import "time"

type PointAccount struct {
	UserID      string    `gorm:"type:varchar(64);primaryKey;column:user_id" json:"user_id"`
	Balance     int64     `gorm:"not null;default:0;column:balance" json:"balance"`
	TotalEarned int64     `gorm:"not null;default:0;column:total_earned" json:"total_earned"`
	UpdatedAt   time.Time `gorm:"not null;column:updated_at" json:"updated_at"`
}

func (PointAccount) TableName() string {
	return "point_accounts"
}

type PointTransaction struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EventID      string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_point_transactions_event;column:event_id" json:"event_id"`
	UserID       string    `gorm:"type:varchar(64);not null;index:idx_point_transactions_user;column:user_id" json:"user_id"`
	Source       string    `gorm:"type:varchar(32);not null;column:source" json:"source"`
	Title        string    `gorm:"type:varchar(255);not null;column:title" json:"title"`
	Points       int64     `gorm:"not null;column:points" json:"points"`
	BalanceAfter int64     `gorm:"not null;column:balance_after" json:"balance_after"`
	OccurredAt   time.Time `gorm:"not null;column:occurred_at" json:"occurred_at"`
	CreatedAt    time.Time `gorm:"not null;column:created_at" json:"created_at"`
}

func (PointTransaction) TableName() string {
	return "point_transactions"
}
