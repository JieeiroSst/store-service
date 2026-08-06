package model

type Category struct {
	ID   int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"column:name"`
}

func (Category) TableName() string {
	return "book_store_category"
}
