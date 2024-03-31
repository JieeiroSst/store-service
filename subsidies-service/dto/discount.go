package dto

type Discount struct {
	ID           int     `json:"id" db:"id"`
	ProgramID    int     `json:"program_id" db:"program_id"`
	ProductID    int     `json:"product_id" db:"product_id"`
	DiscountType string  `json:"discount_type" db:"discount_type"`
	Amount       float64 `json:"amount" db:"amount"`
	SKU          string  `json:"sku" db:"sku"`
	StartDate    string  `json:"start_date" db:"start_date"`
	EndDate      string  `json:"end_date" db:"end_date"`
}
