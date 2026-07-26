package model

type Order struct {
	ID              int     `json:"id" gorm:"column:id;primaryKey"`
	Number          string  `json:"number" gorm:"column:number"`
	Currency        string  `json:"currency" gorm:"column:currency"`
	TotalInclTax    float64 `json:"total_incl_tax" gorm:"column:total_incl_tax"`
	TotalExclTax    float64 `json:"total_excl_tax" gorm:"column:total_excl_tax"`
	ShippingInclTax float64 `json:"shipping_incl_tax" gorm:"column:shipping_incl_tax"`
	ShippingExclTax float64 `json:"shipping_excl_tax" gorm:"column:shipping_excl_tax"`
	ShippingMethod  string  `json:"shipping_method" gorm:"column:shipping_method"`
	ShippingCode    string  `json:"shipping_code" gorm:"column:shipping_code"`
	Status          string  `json:"status" gorm:"column:status"`
	GuestEmail      string  `json:"guest_email" gorm:"column:guest_email"`
	SiteID          int     `json:"site_id" gorm:"column:site_id"`
	UserID          int     `json:"user_id" gorm:"column:user_id"`
	OwnerID         int     `json:"owner_id" gorm:"column:owner_id"`
}

func (Order) TableName() string {
	return "order_order"
}
