package http

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/repository"
)

type inMemoryOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]*entity.Order
}

func NewInMemoryOrderRepository() repository.OrderRepository {
	return &inMemoryOrderRepo{
		orders: make(map[string]*entity.Order),
	}
}

func (r *inMemoryOrderRepo) Create(ctx context.Context, order *entity.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	r.orders[order.ID] = order
	return nil
}

func (r *inMemoryOrderRepo) GetByID(ctx context.Context, orderID string) (*entity.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}
	return order, nil
}

func (r *inMemoryOrderRepo) UpdateStatus(ctx context.Context, orderID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, ok := r.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}
	order.Status = status
	order.UpdatedAt = time.Now()
	return nil
}

func (r *inMemoryOrderRepo) GetStaleOrders(ctx context.Context, olderThanMinutes int) ([]entity.StaleOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cutoff := time.Now().Add(-time.Duration(olderThanMinutes) * time.Minute)
	var stale []entity.StaleOrder
	for _, o := range r.orders {
		if o.Status == "PENDING" && o.CreatedAt.Before(cutoff) {
			stale = append(stale, entity.StaleOrder{
				OrderID:   o.ID,
				Status:    o.Status,
				CreatedAt: o.CreatedAt,
			})
		}
	}
	return stale, nil
}

func (r *inMemoryOrderRepo) Delete(ctx context.Context, orderID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.orders, orderID)
	return nil
}
