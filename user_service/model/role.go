package model

type Role struct {
	Id    int     `gorm:"primaryKey" json:"id"`
	Name  string  `json:"name"`
	Users []Users `gorm:"many2many:user_roles;"`
}
