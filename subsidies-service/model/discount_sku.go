package model

type DiscountBySKU struct {
	Discount
	SKU string `db:"sku"`
}
