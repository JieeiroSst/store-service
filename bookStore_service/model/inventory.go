package model

import "time"

type Inventory struct {
	ID       int       `json:"id"`
	Date     time.Time `json:"date"`
	BookID   int       `json:"book_id"`
	Quantity int       `json:"quantity"`
}
