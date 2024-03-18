package model

type Program struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	StartDate   string `db:"start_date"`
	EndDate     string `db:"end_date"`
}
