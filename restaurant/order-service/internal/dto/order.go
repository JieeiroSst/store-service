package dto

import (
	"time"

	"github.com/JIeeiroSst/order-service/comon"
	"github.com/JIeeiroSst/order-service/internal/model"
	"github.com/JieeiroSst/logger"
)

type Order struct {
	ID          int       `json:"id,omitempty" form:"id"`
	TableName   string    `json:"table_name,omitempty" form:"table_name"`
	Status      string    `json:"status,omitempty" form:"status"`
	KitchenID   int       `json:"kitchen_id,omitempty"`
	TotalAmount float64   `json:"total_amount,omitempty"`
	CreatedDate time.Time `json:"created_date,omitempty"`
	UpdatedTime time.Time `json:"updated_time,omitempty"`
}

func (order Order) CreateOrder() model.Order {
	return model.Order{
		ID:          logger.GearedIntID(),
		TableName:   order.TableName,
		Status:      comon.PendingStatus,
		KitchenID:   order.KitchenID,
		TotalAmount: order.TotalAmount,
		CreatedDate: time.Now(),
		UpdatedTime: time.Now(),
	}
}

func (order Order) CancelOrder() model.Order {
	return model.Order{
		TableName:   order.TableName,
		Status:      comon.CancelStatus,
		KitchenID:   order.KitchenID,
		UpdatedTime: time.Now(),
	}
}

func (order Order) SuccessOrder() model.Order {
	return model.Order{
		TableName:   order.TableName,
		Status:      comon.SuccessStatus,
		KitchenID:   order.KitchenID,
		UpdatedTime: time.Now(),
	}
}

func BuildOrder(order *model.Order) *Order {
	if order == nil {
		return nil
	}
	return &Order{
		ID:          order.ID,
		TableName:   order.TableName,
		Status:      order.Status,
		KitchenID:   order.KitchenID,
		TotalAmount: order.TotalAmount,
		CreatedDate: order.CreatedDate,
		UpdatedTime: order.UpdatedTime,
	}
}
