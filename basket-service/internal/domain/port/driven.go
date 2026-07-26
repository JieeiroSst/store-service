package port

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
)

// BasketRepository persists baskets owned by this service.
type BasketRepository interface {
	Create(ctx context.Context, basket *model.Basket) error
	GetByID(ctx context.Context, id int) (*model.Basket, error)
	Update(ctx context.Context, basket *model.Basket) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Basket, error)
}

// BasketLineRepository persists basket lines owned by this service.
type BasketLineRepository interface {
	Create(ctx context.Context, line *model.BasketLine) error
	GetByID(ctx context.Context, id int) (*model.BasketLine, error)
	Update(ctx context.Context, line *model.BasketLine) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.BasketLine, error)
}

// BasketLineAttributeRepository persists basket line attributes owned by this service.
type BasketLineAttributeRepository interface {
	Create(ctx context.Context, attribute *model.BasketLineAttribute) error
	GetByID(ctx context.Context, id int) (*model.BasketLineAttribute, error)
	Update(ctx context.Context, attribute *model.BasketLineAttribute) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.BasketLineAttribute, error)
}

// OrderRepository reads orders owned by order-processing-service; basket-service
// only references them and must not write to this table.
type OrderRepository interface {
	GetByID(ctx context.Context, id int) (*model.Order, error)
	List(ctx context.Context) ([]model.Order, error)
}

// UserRepository reads users owned by user_service; basket-service only
// references them and must not write to this table.
type UserRepository interface {
	GetByID(ctx context.Context, id int) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
}
