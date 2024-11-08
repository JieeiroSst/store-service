package dto

import (
	"github.com/JIeeiroSst/accounting-service/model"
)

type Order struct {
	ID        int    `json:"id" form:"id"`
	TableName string `json:"table_name" form:"table_name"`
	Status    string `json:"status" form:"status"`
	KitchenID int    `json:"kitchen_id"`
}

type Delivery struct {
	Name      string `json:"name" form:"name"`
	Address   string `json:"address" form:"address"`
	KitchenID int    `json:"kitchen_id"`
}

type AuthCart struct {
	Order    Order    `json:"order" form:"order"`
	Delivery Delivery `json:"delivery" form:"delivery"`
}

func (o *Order) Build(status string) model.Order {
	return model.Order{
		ID:        o.ID,
		TableName: o.TableName,
		Status:    status,
		KitchenID: o.KitchenID,
	}
}

func (d *Delivery) Build() model.Delivery {
	return model.Delivery{
		Name:      d.Name,
		Address:   d.Address,
		KitchenID: d.KitchenID,
	}
}
