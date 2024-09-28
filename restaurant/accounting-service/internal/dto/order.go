package dto

type Order struct {
	ID        int    `json:"id" form:"id"`
	TableName string `json:"table_name" form:"table_name"`
	Status    string `json:"status" form:"status"`
	KitchenID int    `json:"kitchen_id"`
}
