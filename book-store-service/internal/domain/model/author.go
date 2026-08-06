package model

type Author struct {
	ID   int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"column:name"`
	Bio  string `json:"bio" gorm:"column:bio"`
}

func (Author) TableName() string {
	return "book_store_author"
}
