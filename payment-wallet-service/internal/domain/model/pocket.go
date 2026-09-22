package model

import "time"

type Pocket struct {
	PocketID  string    `gorm:"column:pocket_id;primaryKey" json:"pocket_id"`
	WalletID  string    `gorm:"column:wallet_id" json:"wallet_id"`
	Name      string    `gorm:"column:name" json:"name"`
	Balance   int64     `gorm:"column:balance" json:"balance"`
	Currency  string    `gorm:"column:currency" json:"currency"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Pocket) TableName() string { return "pockets" }
