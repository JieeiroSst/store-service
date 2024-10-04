package dto

type Order struct {
	ID        int    `json:"id" form:"id"`
	TableName string `json:"table_name" form:"table_name"`
	Status    string `json:"status" form:"status"`
	KitchenID int    `json:"kitchen_id"`
}

type Delivery struct {
	ShipID  int    `json:"ship_id" form:"ship_id"`
	Name    string `json:"name" form:"name"`
	Address string `json:"address" form:"address"`
}

type AuthCart struct {
	Order    Order    `json:"order" form:"order"`
	Delivery Delivery `json:"delivery" form:"delivery"`
}
