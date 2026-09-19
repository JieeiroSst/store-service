package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/order-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	Update(ctx context.Context, id int, order *model.Order) error
	FindByID(ctx context.Context, id int) (*model.Order, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

// FoodPrice is one food's authoritative name/price, as looked up from
// kitchen-service's catalog.
type FoodPrice struct {
	ID    int
	Name  string
	Price float64
}

// FoodPricer looks up authoritative food prices from kitchen-service, so
// an Order's TotalAmount is computed here rather than trusted from a caller.
type FoodPricer interface {
	GetPrices(ctx context.Context, foodIDs []int) (map[int]FoodPrice, error)
}

type ReservationRepository interface {
	Create(ctx context.Context, reservation model.Reservation) error
	Cancel(ctx context.Context, id int) error
	FindByID(ctx context.Context, id int) (*model.Reservation, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	// FindConflicting returns bookings for the same table within window of
	// at, used to reject double-booking a table.
	FindConflicting(ctx context.Context, tableName string, at time.Time, window time.Duration) ([]model.Reservation, error)
}
