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
	ListByUser(ctx context.Context, userID int) ([]model.Basket, error)
	ListAwaitingFulfillment(ctx context.Context, limit int) ([]model.Basket, error)
}

// BasketLineRepository persists basket lines owned by this service.
type BasketLineRepository interface {
	Create(ctx context.Context, line *model.BasketLine) error
	GetByID(ctx context.Context, id int) (*model.BasketLine, error)
	Update(ctx context.Context, line *model.BasketLine) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.BasketLine, error)
	ListByBasket(ctx context.Context, basketID int) ([]model.BasketLine, error)
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
	ListByUser(ctx context.Context, userID int) ([]model.Order, error)
}

type UserDirectory interface {
	GetByID(ctx context.Context, id int) (*model.User, error)
}

type StockRecord struct {
	ID            int
	ProductID     int
	PriceCurrency string
	Price         float64
	NumAvailable  *int
}

func (s StockRecord) CanSupply(qty int) bool { return s.NumAvailable == nil || *s.NumAvailable >= qty }

type Catalogue interface {
	StockRecord(ctx context.Context, id int) (*StockRecord, error)
	Allocate(ctx context.Context, id, quantity int) error
	Release(ctx context.Context, id, quantity int) error
}

type WalletPayment struct {
	UserID         int
	WalletID       string
	Amount         float64
	Pin            string
	Description    string
	IdempotencyKey string
	Metadata       map[string]string
}

type Wallet interface {
	Pay(ctx context.Context, p WalletPayment) (string, error)
}

type PaidOrderLine struct {
	ProductID     int     `json:"product_id"`
	StockrecordID int     `json:"stockrecord_id"`
	Title         string  `json:"title"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
}

type PaidOrder struct {
	BasketID          int             `json:"basket_id"`
	UserID            int             `json:"user_id"`
	Currency          string          `json:"currency"`
	Total             float64         `json:"total"`
	PaymentReference  string          `json:"payment_reference"`
	ShippingAddressID int             `json:"shipping_address_id"`
	Lines             []PaidOrderLine `json:"lines"`
}

type Fulfillment interface {
	StartPaidOrder(ctx context.Context, order PaidOrder) (string, error)
}
