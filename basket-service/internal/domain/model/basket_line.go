package model

type BasketLine struct {
	ID            int     `json:"id" gorm:"column:id;primaryKey"`
	LineReference string  `json:"line_reference" gorm:"column:line_reference"`
	Quantity      int     `json:"quantity" gorm:"column:quantity"`
	PriceCurrency string  `json:"price_currency" gorm:"column:price_currency"`
	PriceExclTax  float64 `json:"price_excl_tax" gorm:"column:price_excl_tax"`
	PriceInclTax  float64 `json:"price_incl_tax" gorm:"column:price_incl_tax"`
	BasketID      int     `json:"basket_id" gorm:"column:basket_id"`
	ProductID     int     `json:"product_id" gorm:"column:product_id"`
	StockrecordID int     `json:"stockrecord_id" gorm:"column:stockrecord_id"`
}

func (BasketLine) TableName() string {
	return "basket_line"
}
