package model

type Consumer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	KitchenID int    `json:"kitchen_id"`
	Menu      string `json:"menu" form:"menu"`
}
