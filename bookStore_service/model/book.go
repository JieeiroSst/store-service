package model

type Book struct {
	ID                int `gorm:"primary_key"`
	Name              string
	Author            string
	Publisher         string
	BookType          string
	Price             float64
	Quantity          int
	UnitOfMeasurement string
}
