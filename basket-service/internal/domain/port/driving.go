package port

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
)

type BasketUsecase interface {
	CreateBasket(ctx context.Context, basket *model.Basket) (*model.Basket, error)
	GetBasket(ctx context.Context, id int) (*model.Basket, error)
	ListBaskets(ctx context.Context) ([]model.Basket, error)
	ListBasketsByUser(ctx context.Context, userID int) ([]model.Basket, error)
	UpdateBasket(ctx context.Context, basket *model.Basket) (*model.Basket, error)
	DeleteBasket(ctx context.Context, id int) error
}

type BasketLineUsecase interface {
	CreateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error)
	GetBasketLine(ctx context.Context, id int) (*model.BasketLine, error)
	ListBasketLines(ctx context.Context) ([]model.BasketLine, error)
	UpdateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error)
	DeleteBasketLine(ctx context.Context, id int) error
	ListBasketLinesByBasket(ctx context.Context, basketID int) ([]model.BasketLine, error)
}

type BasketLineAttributeUsecase interface {
	CreateBasketLineAttribute(ctx context.Context, attribute *model.BasketLineAttribute) (*model.BasketLineAttribute, error)
	GetBasketLineAttribute(ctx context.Context, id int) (*model.BasketLineAttribute, error)
	ListBasketLineAttributes(ctx context.Context) ([]model.BasketLineAttribute, error)
	UpdateBasketLineAttribute(ctx context.Context, attribute *model.BasketLineAttribute) (*model.BasketLineAttribute, error)
	DeleteBasketLineAttribute(ctx context.Context, id int) error
}

type CheckoutCommand struct {
	BasketID          int
	UserID            int
	WalletID          string
	Pin               string
	ShippingAddressID int
	AccessToken       string
	ShippingFee       *float64
}

type ShippingQuoteCommand struct {
	BasketID          int
	UserID            int
	ShippingAddressID int
	AccessToken       string
}

type ShippingQuoteResult struct {
	Currency           string  `json:"currency"`
	Subtotal           float64 `json:"subtotal"`
	ShippingFee        float64 `json:"shipping_fee"`
	Total              float64 `json:"total"`
	ShippingCharged    bool    `json:"shipping_charged"`
	ServiceName        string  `json:"service_name,omitempty"`
	ExpectedDeliveryAt string  `json:"expected_delivery_at,omitempty"`
}

type CheckoutUsecase interface {
	Checkout(ctx context.Context, cmd CheckoutCommand) (*model.Basket, error)
	QuoteShipping(ctx context.Context, cmd ShippingQuoteCommand) (*ShippingQuoteResult, error)
	RetryFulfillment(ctx context.Context, limit int) (int, error)
}

type OrderUsecase interface {
	GetOrder(ctx context.Context, id int) (*model.Order, error)
	ListOrders(ctx context.Context) ([]model.Order, error)
	ListOrdersByUser(ctx context.Context, userID int) ([]model.Order, error)
}

type UserUsecase interface {
	GetUser(ctx context.Context, id int) (*model.User, error)
}
