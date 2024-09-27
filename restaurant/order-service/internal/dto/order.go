package dto

type Order struct {
	ID        int    `json:"id"`
	TableName string `json:"table_name"`
	Status    string `json:"status"`
	MenuID    string `json:"menu_id"`
}
