package model

type Basket struct {
	ID                int    `json:"id" gorm:"column:id;primaryKey"`
	Status            string `json:"status" gorm:"column:status"`
	DatePlaced        string `json:"date_placed" gorm:"column:date_placed"`
	DateCreated       string `json:"date_created" gorm:"column:date_created"`
	DateMerged        string `json:"date_merged" gorm:"column:date_merged"`
	DateSubmitted     string `json:"date_submitted" gorm:"column:date_submitted"`
	BillingAddressID  int    `json:"billing_address_id" gorm:"column:billing_address_id"`
	ShippingAddressID int    `json:"shipping_address_id" gorm:"column:shipping_address_id"`
	UserID            int    `json:"user_id" gorm:"column:user_id"`
}

func (Basket) TableName() string {
	return "basket_basket"
}
