package model

type OrderBillingaddress struct {
	ID             int            `json:"id,omitempty"`
	Title          string         `json:"title,omitempty"`
	FirstName      string         `json:"first_name,omitempty"`
	LastName       string         `json:"last_name,omitempty"`
	Line1          string         `json:"line1,omitempty"`
	Line2          string         `json:"line2,omitempty"`
	Line3          string         `json:"line3,omitempty"`
	Line4          string         `json:"line4,omitempty"`
	State          string         `json:"state,omitempty"`
	Postcode       string         `json:"postcode,omitempty"`
	SearchText     string         `json:"search_text,omitempty"`
	CountryID      string         `json:"country_id,omitempty"`
	AddressCountry AddressCountry `json:"address_country,omitempty" gorm:"foreignKey:CountryID"`
}
