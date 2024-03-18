package model

type Discount struct {
	ID           int     `db:"id"`
	ProgramID    int     `db:"program_id"`
	ProductID    int     `db:"product_id"`
	DiscountType string  `db:"discount_type"`
	Amount       float64 `db:"amount"`
	StartDate    string  `db:"start_date"`
	EndDate      string  `db:"end_date"`
}
