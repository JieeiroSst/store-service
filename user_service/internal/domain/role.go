package domain

type Role struct {
	Id    int    `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Users []User `gorm:"many2many:user_roles;"`
}
