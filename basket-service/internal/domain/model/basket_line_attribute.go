package model

type BasketLineAttribute struct {
	ID       int    `json:"id" gorm:"column:id;primaryKey"`
	BasketID int    `json:"basket_id" gorm:"column:basket_id"`
	LineID   int    `json:"line_id" gorm:"column:line_id"`
	OptionID int    `json:"option_id" gorm:"column:option_id"`
	Value    string `json:"value" gorm:"column:value"`
}

func (BasketLineAttribute) TableName() string {
	return "basket_basket_vouchers"
}
