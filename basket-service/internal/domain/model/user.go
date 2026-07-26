package model

type User struct {
	ID          int    `json:"id" gorm:"column:id;primaryKey"`
	Username    string `json:"username" gorm:"column:username"`
	Password    string `json:"-" gorm:"column:password"`
	FirstName   string `json:"first_name" gorm:"column:first_name"`
	LastName    string `json:"last_name" gorm:"column:last_name"`
	Email       string `json:"email" gorm:"column:email"`
	IsSuperuser bool   `json:"is_superuser" gorm:"column:is_superuser"`
	IsStaff     bool   `json:"is_staff" gorm:"column:is_staff"`
	IsActive    bool   `json:"is_active" gorm:"column:is_active"`
	DateJoined  string `json:"date_joined" gorm:"column:date_joined"`
	LastLogin   string `json:"last_login" gorm:"column:last_login"`
}

func (User) TableName() string {
	return "auth_user"
}
