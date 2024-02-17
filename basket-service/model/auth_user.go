package model

type User struct {
	ID          int    `db:"id"`
	Username    string `db:"username"`
	Password    string `db:"password"`
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	Email       string `db:"email"`
	IsSuperuser bool   `db:"is_superuser"`
	IsStaff     bool   `db:"is_staff"`
	IsActive    bool   `db:"is_active"`
	DateJoined  string `db:"date_joined"`
	LastLogin   string `db:"last_login"`
}
