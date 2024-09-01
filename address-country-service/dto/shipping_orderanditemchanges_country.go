package dto

type ShippingOrderanditemchangesCountry struct {
	ID                    int    `json:"id,omitempty"`
	OrderanditemchangesID int    `json:"orderanditemchanges_id,omitempty"`
	CountryID             string `json:"country_id,omitempty"`
}
