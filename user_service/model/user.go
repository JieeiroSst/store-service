package model

import "time"

type Users struct {
	Id         int       `gorm:"primaryKey" json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	Sex        string    `json:"sex"`
	Checked    bool      `json:"checked"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"default:null"`
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
