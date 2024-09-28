package dto

import (
	"github.com/JIeeiroSst/order-service/comon"
	"github.com/JIeeiroSst/order-service/internal/model"
	"github.com/JieeiroSst/logger"
)

type Order struct {
	ID        int    `json:"id" form:"id"`
	TableName string `json:"table_name" form:"table_name"`
	Status    string `json:"status" form:"status"`
	KitchenID int    `json:"kitchen_id"`
}

func (order Order) CreateOrder() model.Order {
	return model.Order{
		ID:        logger.GearedIntID(),
		TableName: order.TableName,
		Status:    comon.PendingStatus,
		KitchenID: order.KitchenID,
	}
}

func (order Order) CancelOrder() model.Order {
	return model.Order{
		TableName: order.TableName,
		Status:    comon.CancelStatus,
		KitchenID: order.KitchenID,
	}
}

func (order Order) SuccessOrder() model.Order {
	return model.Order{
		TableName: order.TableName,
		Status:    comon.SuccessStatus,
		KitchenID: order.KitchenID,
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
		KitchenID: order.KitchenID,
	}
}
