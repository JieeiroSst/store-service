package dto

type Delivery struct {
	ShipID    int    `json:"ship_id" form:"ship_id"`
	Name      string `json:"name" form:"name"`
	Address   string `json:"address" form:"address"`
	KitchenID int    `json:"kitchen_id"`
}
