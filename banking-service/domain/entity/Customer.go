package entity

type Customer struct {
	CustomerID   int    `json:"customer_id" gorm:"column:customer_id"`
	CustomerType string `json:"customer_type" gorm:"column:customer_type"`
}

func (Customer) TableName() string {
	return "customers"
}