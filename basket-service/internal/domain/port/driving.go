package port

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
)

type BasketUsecase interface {
	CreateBasket(ctx context.Context, basket *model.Basket) (*model.Basket, error)
	GetBasket(ctx context.Context, id int) (*model.Basket, error)
	ListBaskets(ctx context.Context) ([]model.Basket, error)
	UpdateBasket(ctx context.Context, basket *model.Basket) (*model.Basket, error)
	DeleteBasket(ctx context.Context, id int) error
}

type BasketLineUsecase interface {
	CreateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error)
	GetBasketLine(ctx context.Context, id int) (*model.BasketLine, error)
	ListBasketLines(ctx context.Context) ([]model.BasketLine, error)
	UpdateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error)
	DeleteBasketLine(ctx context.Context, id int) error
}

type BasketLineAttributeUsecase interface {
	CreateBasketLineAttribute(ctx context.Context, attribute *model.BasketLineAttribute) (*model.BasketLineAttribute, error)
	GetBasketLineAttribute(ctx context.Context, id int) (*model.BasketLineAttribute, error)
	ListBasketLineAttributes(ctx context.Context) ([]model.BasketLineAttribute, error)
	UpdateBasketLineAttribute(ctx context.Context, attribute *model.BasketLineAttribute) (*model.BasketLineAttribute, error)
	DeleteBasketLineAttribute(ctx context.Context, id int) error
}

type OrderUsecase interface {
	GetOrder(ctx context.Context, id int) (*model.Order, error)
	ListOrders(ctx context.Context) ([]model.Order, error)
}

type UserUsecase interface {
	GetUser(ctx context.Context, id int) (*model.User, error)
	ListUsers(ctx context.Context) ([]model.User, error)
}
