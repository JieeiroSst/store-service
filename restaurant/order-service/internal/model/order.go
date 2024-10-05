package model

import "time"

type Order struct {
	ID          int       `json:"id"`
	TableName   string    `json:"table_name"`
	Status      string    `json:"status"`
	KitchenID   int       `json:"kitchen_id"`
	TotalAmount float64   `json:"total_amount"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedTime time.Time `json:"updated_time"`
}
