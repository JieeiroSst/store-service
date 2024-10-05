package dto

import (
	"github.com/JIeeiroSst/delivery-service/internal/model"
	"github.com/JieeiroSst/logger"
)

type Delivery struct {
	ShipID    int    `json:"ship_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address,omitempty"`
	KitchenID int    `json:"kitchen_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

type AddressDelivery struct {
	Name      string `json:"name" form:"name"`
	Address   string `json:"address" form:"address"`
	KitchenID int    `json:"kitchen_id"`
}

func BuildUpdate(d *Delivery, a AddressDelivery) *Delivery {
	if d == nil {
		return nil
	}
	return &Delivery{
		ShipID:    d.ShipID,
		Name:      a.Name,
		Address:   a.Address,
		KitchenID: a.KitchenID,
		Status:    "busy",
	}
}

func (d Delivery) Create() model.Delivery {
	return model.Delivery{
		ShipID:  logger.GearedIntID(),
		Name:    d.Name,
		Address: d.Address,
		Status:  1,
	}
}

func BuildDtoStatus(status string) int {
	if status == "busy" {
		return 0
	}
	return 1
}

func BuildStatus(status int) string {
	if status == 0 {
		return "busy"
	}
	return "free"
}

func (d Delivery) Update() model.Delivery {
	status := BuildDtoStatus(d.Status)
	return model.Delivery{
		Name:      d.Name,
		Address:   d.Address,
		KitchenID: d.KitchenID,
		Status:    status,
	}
}

func BuildDelivery(delivery *model.Delivery) *Delivery {
	if delivery == nil {
		return nil
	}
	status := BuildStatus(delivery.Status)
	return &Delivery{
		ShipID:    delivery.ShipID,
		Name:      delivery.Name,
		Address:   delivery.Address,
		KitchenID: delivery.KitchenID,
		Status:    status,
	}
}
