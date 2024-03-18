package dto

type DiscountBySKU struct {
	Discount
	SKU string `db:"sku"`
}
