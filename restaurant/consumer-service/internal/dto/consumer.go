package dto

import (
	"strconv"
	"strings"

	"github.com/JIeeiroSst/consumer-service/internal/model"
)

type Consumer struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	KitchenID int        `json:"kitchen_id"`
	Menu      []MenuFood `json:"menu" form:"menu"`
	OrderID   int        `json:"order_id"`
}

type MenuFood struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Category Category `json:"category"`
}
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (d Consumer) Build() model.Consumer {
	menu := make([]string, 0)
	for _, v := range d.Menu {
		menu = append(menu, strconv.Itoa(v.ID))
	}
	return model.Consumer{
		ID:        d.ID,
		Name:      d.Name,
		KitchenID: d.KitchenID,
		Menu:      strings.Join(menu, ","),
	}
}

type Order struct {
	ID        int    `json:"id" form:"id"`
	TableName string `json:"table_name" form:"table_name"`
	KitchenID int    `json:"kitchen_id"`
}

func (c Consumer) BuildV2() Order {
	return Order{
		ID:        c.OrderID,
		TableName: c.Name,
		KitchenID: c.KitchenID,
	}
}
