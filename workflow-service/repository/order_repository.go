package repository

import (
	"errors"
	"sync"

	"github.com/JIeeiroSst/workflow-service/app"
)

type InMemoryOrderRepository struct {
	orders map[string]*app.Order
	mu     sync.RWMutex
}

func NewOrderRepository(orders map[string]*app.Order) *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: orders,
	}
}

func (r *InMemoryOrderRepository) GetOrder(orderID string) (*app.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (r *InMemoryOrderRepository) UpdateOrder(order *app.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if order.ID == "" {
		return errors.New("order ID cannot be empty")
	}

	r.orders[order.ID] = order
	return nil
}
