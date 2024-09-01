package dto

import "time"

type UserAddress struct {
	ID                            int       `json:"id,omitempty"`
	Title                         string    `json:"title,omitempty"`
	FirstName                     string    `json:"first_name,omitempty"`
	LastName                      string    `json:"last_name,omitempty"`
	Line1                         string    `json:"line1,omitempty"`
	Line2                         string    `json:"line2,omitempty"`
	Line3                         string    `json:"line3,omitempty"`
	Line4                         string    `json:"line4,omitempty"`
	State                         string    `json:"state,omitempty"`
	Postcode                      string    `json:"postcode,omitempty"`
	SearchText                    string    `json:"search_text,omitempty"`
	CountryID                     string    `json:"country_id,omitempty"`
	PhoneNumber                   string    `json:"phone_number,omitempty"`
	Notes                         string    `json:"notes,omitempty"`
	IsDefaultForShipping          bool      `json:"is_default_for_shipping,omitempty"`
	IsDefaultForBilling           bool      `json:"is_default_for_billing,omitempty"`
	NumberOrdersAsShippingAddress int       `json:"number_orders_as_shipping_address,omitempty"`
	Hash                          string    `json:"hash,omitempty"`
	DateCreated                   time.Time `json:"date_created,omitempty"`
	UserID                        int       `json:"user_id,omitempty"`
	NumberOrdersAsBillingAddress  int       `json:"number_orders_as_billing_address,omitempty"`
}
