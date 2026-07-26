package port

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
)

// AddressRepository is the secondary port for address persistence.
type AddressRepository interface {
	Create(ctx context.Context, address *model.Address) error
	GetByID(ctx context.Context, id int) (*model.Address, error)
	Update(ctx context.Context, address *model.Address) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Address, error)
}

// CustomerRepository is the secondary port for customer persistence.
type CustomerRepository interface {
	Create(ctx context.Context, customer *model.Customer) error
	GetByID(ctx context.Context, id int) (*model.Customer, error)
	Update(ctx context.Context, customer *model.Customer) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Customer, error)
}

// PlanRepository is the secondary port for plan persistence.
type PlanRepository interface {
	Create(ctx context.Context, plan *model.Plan) error
	GetByID(ctx context.Context, id int) (*model.Plan, error)
	Update(ctx context.Context, plan *model.Plan) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Plan, error)
}

// SubscriptionRepository is the secondary port for subscription persistence.
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *model.Subscription) error
	GetByID(ctx context.Context, id int) (*model.Subscription, error)
	Update(ctx context.Context, subscription *model.Subscription) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Subscription, error)
}

// InvoiceRepository is the secondary port for invoice persistence.
type InvoiceRepository interface {
	Create(ctx context.Context, invoice *model.Invoice) error
	GetByID(ctx context.Context, id int) (*model.Invoice, error)
	Update(ctx context.Context, invoice *model.Invoice) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Invoice, error)
}

// TransactionRepository is the secondary port for transaction persistence.
type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) error
	GetByID(ctx context.Context, id int) (*model.Transaction, error)
	Update(ctx context.Context, transaction *model.Transaction) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Transaction, error)
}
