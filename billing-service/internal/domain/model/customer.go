package model

type Customer struct {
	CustomerID  int    `gorm:"column:customer_id;primaryKey;autoIncrement" json:"customer_id,omitempty"`
	Name        string `gorm:"column:name" json:"name"`
	Company     string `gorm:"column:company" json:"company,omitempty"`
	Email       string `gorm:"column:email" json:"email"`
	PhoneNumber string `gorm:"column:phone_number" json:"phone_number"`
	AddressID   int    `gorm:"column:address_id" json:"address_id,omitempty"`
}

func (Customer) TableName() string { return "customers" }
