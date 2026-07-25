package model

import "time"

type Account struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateCreated time.Time `json:"date_created"`
}

func (Account) TableName() string { return "accounts" }
