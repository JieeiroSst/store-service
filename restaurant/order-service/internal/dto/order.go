package dto

import (
	"strings"

	"github.com/JIeeiroSst/order-service/comon"
	"github.com/JIeeiroSst/order-service/internal/model"
	"github.com/JieeiroSst/logger"
)

type Order struct {
	ID        int      `json:"id"`
	TableName string   `json:"table_name"`
	Status    string   `json:"status"`
	MenuIDs   []string `json:"menu_id"`
}

func (order Order) CreateOrder() model.Order {
	return model.Order{
		ID:        logger.GearedIntID(),
		TableName: order.TableName,
		Status:    comon.PendingStatus,
		MenuID:    strings.Join(order.MenuIDs, ","),
	}
}

func (order Order) CancelOrder() model.Order {
	return model.Order{
		TableName: order.TableName,
		Status:    comon.CancelStatus,
		MenuID:    strings.Join(order.MenuIDs, ","),
	}
}

func (order Order) SuccessOrder() model.Order {
	return model.Order{
		TableName: order.TableName,
		Status:    comon.SuccessStatus,
		MenuID:    strings.Join(order.MenuIDs, ","),
	}
}

func BuildOrder(order *model.Order) *Order {
	if order == nil {
		return nil
	}
	return &Order{
		ID:        order.ID,
		TableName: order.TableName,
		Status:    order.Status,
		MenuIDs:   strings.Split(order.MenuID, ","),
	}
}
