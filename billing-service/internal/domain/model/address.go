package model

type Address struct {
	AddressID  int    `gorm:"column:address_id;primaryKey;autoIncrement" json:"address_id,omitempty"`
	Line1      string `gorm:"column:line1" json:"line1"`
	Line2      string `gorm:"column:line2" json:"line2,omitempty"`
	City       string `gorm:"column:city" json:"city"`
	State      string `gorm:"column:state" json:"state"`
	PostalCode string `gorm:"column:postal_code" json:"postal_code"`
	Country    string `gorm:"column:country" json:"country"`
}

func (Address) TableName() string { return "address" }
