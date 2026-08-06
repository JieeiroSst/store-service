package model

type Publisher struct {
	ID      int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name    string `json:"name" gorm:"column:name"`
	Address string `json:"address" gorm:"column:address"`
}

func (Publisher) TableName() string {
	return "book_store_publisher"
}
