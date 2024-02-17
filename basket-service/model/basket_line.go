package model

type BasketLine struct {
	ID            int     `db:"id"`
	LineReference string  `db:"line_reference"`
	Quantity      int     `db:"quantity"`
	PriceCurrency string  `db:"price_currency"`
	PriceExclTax  float64 `db:"price_excl_tax"`
	PriceInclTax  float64 `db:"price_incl_tax"`
	BasketID      int     `db:"basket_id"`
	ProductID     int     `db:"product_id"`
	StockrecordID int     `db:"stockrecord_id"`
}
