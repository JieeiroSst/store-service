package model

type Delivery struct {
	ShipID    int    `json:"ship_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address,omitempty"`
	KitchenID int    `json:"kitchen_id,omitempty"`
	Status    int    `json:"status,omitempty"`
}
