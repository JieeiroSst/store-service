package dto

type Program struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	StartDate   string     `json:"start_date" db:"start_date"`
	EndDate     string     `json:"end_date" db:"end_date"`
	Discounts   []Discount `json:"discounts"`
}

type ProgramPage struct {
	Programs []Program `json:"programs"`
	HasNext  bool      `json:"has_next"`
}
