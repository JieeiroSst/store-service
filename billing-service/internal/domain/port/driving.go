package port

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
)

// AddressUsecase is the primary port for address operations.
type AddressUsecase interface {
	Create(ctx context.Context, address *model.Address) error
	Get(ctx context.Context, id int) (*model.Address, error)
	Update(ctx context.Context, address *model.Address) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Address, error)
}

// CustomerUsecase is the primary port for customer operations.
type CustomerUsecase interface {
	Create(ctx context.Context, customer *model.Customer) error
	Get(ctx context.Context, id int) (*model.Customer, error)
	Update(ctx context.Context, customer *model.Customer) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Customer, error)
}

// PlanUsecase is the primary port for plan operations.
type PlanUsecase interface {
	Create(ctx context.Context, plan *model.Plan) error
	Get(ctx context.Context, id int) (*model.Plan, error)
	Update(ctx context.Context, plan *model.Plan) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Plan, error)
}

// SubscriptionUsecase is the primary port for subscription operations.
type SubscriptionUsecase interface {
	Create(ctx context.Context, subscription *model.Subscription) error
	Get(ctx context.Context, id int) (*model.Subscription, error)
	Update(ctx context.Context, subscription *model.Subscription) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Subscription, error)
}

// InvoiceUsecase is the primary port for invoice operations.
type InvoiceUsecase interface {
	Create(ctx context.Context, invoice *model.Invoice) error
	Get(ctx context.Context, id int) (*model.Invoice, error)
	Update(ctx context.Context, invoice *model.Invoice) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Invoice, error)
}

// TransactionUsecase is the primary port for transaction operations.
type TransactionUsecase interface {
	Create(ctx context.Context, transaction *model.Transaction) error
	Get(ctx context.Context, id int) (*model.Transaction, error)
	Update(ctx context.Context, transaction *model.Transaction) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Transaction, error)
}
