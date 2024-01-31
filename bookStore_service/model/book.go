package model

type Book struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Author            string  `json:"author"`
	Publisher         string  `json:"publisher"`
	BookType          string  `json:"book_type"`
	Price             float64 `json:"price"`
	Quantity          int     `json:"quantity"`
	UnitOfMeasurement string  `json:"unit_of_measurement"`
}
