package model

type Order struct {
	ID        int    `json:"id"`
	TableName string `json:"table_name"`
	Status    string `json:"status"`
	KitchenID int    `json:"kitchen_id"`
}
