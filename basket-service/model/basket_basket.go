package model

type Basket struct {
	ID                int    `db:"id"`
	Status            string `db:"status"`
	DatePlaced        string `db:"date_placed"`
	DateCreated       string `db:"date_created"`
	DateMerged        string `db:"date_merged"`
	DateSubmitted     string `db:"date_submitted"`
	BillingAddressID  int    `db:"billing_address_id"`
	ShippingAddressID int    `db:"shipping_address_id"`
	UserID            int    `db:"user_id"`
}
