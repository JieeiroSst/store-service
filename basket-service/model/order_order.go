package model

type Order struct {
	ID              int     `db:"id"`
	Number          string  `db:"number"`
	AuthUser        *User   `db:"auth_user"`
	Currency        string  `db:"currency"`
	TotalInclTax    float64 `db:"total_incl_tax"`
	TotalExclTax    float64 `db:"total_excl_tax"`
	ShippingInclTax float64 `db:"shipping_incl_tax"`
	ShippingExclTax float64 `db:"shipping_excl_tax"`
	ShippingMethod  string  `db:"shipping_method"`
	ShippingCode    string  `db:"shipping_code"`
	Status          string  `db:"status"`
	GuestEmail      string  `db:"guest_email"`
	Basket          *Basket `db:"basket"`
	SiteID          int     `db:"site_id"`
	UserID          int     `db:"user_id"`
	OwnerID         int     `db:"owner_id"`
}
