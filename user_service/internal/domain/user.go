package domain

import "time"

type User struct {
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
	Roles      []Role    `gorm:"many2many:user_roles;"`
}
