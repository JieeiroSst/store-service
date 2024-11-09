package model

type ShippingWeightbasedCountry struct {
	ID             int            `json:"id,omitempty"`
	WeightbasedID  int            `json:"weightbased_id,omitempty"`
	CountryID      string         `json:"country_id,omitempty"`
	AddressCountry AddressCountry `json:"address_country,omitempty" gorm:"foreignKey:CountryID"`
}
