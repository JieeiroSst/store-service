package model

type AddressCountry struct {
	Id                string `json:"iso31661a2,omitempty"`
	PrinatableName    string `json:"prinatable_name,omitempty"`
	Name              string `json:"name,omitempty"`
	DisplayOrder      int    `json:"display_order,omitempty"`
	IsShippingCountry bool   `json:"is_shipping_country,omitempty"`
}
