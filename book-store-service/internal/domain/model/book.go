package model

type Book struct {
	ID                int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name              string     `json:"name" gorm:"column:name"`
	AuthorID          int        `json:"author_id" gorm:"column:author_id"`
	Author            *Author    `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	PublisherID       int        `json:"publisher_id" gorm:"column:publisher_id"`
	Publisher         *Publisher `json:"publisher,omitempty" gorm:"foreignKey:PublisherID"`
	CategoryID        int        `json:"category_id" gorm:"column:category_id"`
	Category          *Category  `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Price             float64    `json:"price" gorm:"column:price"`
	Quantity          int        `json:"quantity" gorm:"column:quantity"`
	UnitOfMeasurement string     `json:"unit_of_measurement" gorm:"column:unit_of_measurement"`
	PurchaseCount     int        `json:"purchase_count" gorm:"column:purchase_count"`
	ReadCount         int        `json:"read_count" gorm:"column:read_count"`
}

func (Book) TableName() string {
	return "book_store_book"
}
