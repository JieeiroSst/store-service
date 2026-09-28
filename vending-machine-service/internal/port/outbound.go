package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type MachineRepository interface {
	Create(ctx context.Context, m *domain.Machine) error
	Get(ctx context.Context, id string) (*domain.Machine, error)
	List(ctx context.Context, status string, page domain.PageRequest) (domain.Page[domain.Machine], error)
	Update(ctx context.Context, m *domain.Machine) error
}

type MaintenanceRepository interface {
	Create(ctx context.Context, l *domain.MaintenanceLog) error
	ListByMachine(ctx context.Context, machineID string) ([]domain.MaintenanceLog, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, c *domain.Category) error
	List(ctx context.Context, page domain.PageRequest) (domain.Page[domain.Category], error)
}

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Get(ctx context.Context, id string) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	List(ctx context.Context, categoryID string, page domain.PageRequest) (domain.Page[domain.Product], error)
}

type InventoryRepository interface {
	Create(ctx context.Context, i *domain.Inventory) error
	Update(ctx context.Context, i *domain.Inventory) error
	Get(ctx context.Context, id string) (*domain.Inventory, error)
	GetForUpdate(ctx context.Context, id string) (*domain.Inventory, error)
	GetBySlotForUpdate(ctx context.Context, machineID, slot string) (*domain.Inventory, error)
	ListByMachine(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error)
	ListLow(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error)
}

type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error
	Get(ctx context.Context, id string) (*domain.Session, error)
	GetForUpdate(ctx context.Context, id string) (*domain.Session, error)
	Update(ctx context.Context, s *domain.Session) error
	ListExpired(ctx context.Context, now time.Time, limit int) ([]string, error)
}

type ReservationRepository interface {
	Create(ctx context.Context, r *domain.Reservation) error
	Get(ctx context.Context, id string) (*domain.Reservation, error)
	GetForUpdate(ctx context.Context, id string) (*domain.Reservation, error)
	Update(ctx context.Context, r *domain.Reservation) error
	ListExpiredPending(ctx context.Context, now time.Time, limit int) ([]string, error)
	ListPendingBySession(ctx context.Context, sessionID string) ([]string, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	Get(ctx context.Context, id string) (*domain.Payment, error)
	GetForUpdate(ctx context.Context, id string) (*domain.Payment, error)
	Update(ctx context.Context, p *domain.Payment) error
	ListStalePending(ctx context.Context, before time.Time, limit int) ([]string, error)
}

type OrderRepository interface {
	Create(ctx context.Context, o *domain.Order) error
	Get(ctx context.Context, id string) (*domain.Order, error)
	GetForUpdate(ctx context.Context, id string) (*domain.Order, error)
	Update(ctx context.Context, o *domain.Order) error
	ListBySession(ctx context.Context, sessionID string) ([]domain.Order, error)
	GetByPayment(ctx context.Context, paymentID string) (*domain.Order, error)
}

type ReportRepository interface {
	SalesByProduct(ctx context.Context, machineID string, from, to time.Time) ([]domain.ProductSales, error)
}

type EventRepository interface {
	Append(ctx context.Context, e *domain.Event) error
	ListByMachine(ctx context.Context, machineID string, limit int) ([]domain.Event, error)
	LockUndispatched(ctx context.Context, limit int) ([]domain.Event, error)
	MarkDispatched(ctx context.Context, ids []string, at time.Time) error
}

type ChargeRequest struct {
	PaymentID   string
	AmountCents int
	Currency    string
	Method      domain.PaymentMethod
	WalletID    string
}

type RefundRequest struct {
	PaymentID     string
	TransactionID string
	AmountCents   int
	Method        domain.PaymentMethod
	Reason        string
}

type LookupRequest struct {
	PaymentID string
	Method    domain.PaymentMethod
	WalletID  string
}

type PaymentGateway interface {
	Charge(ctx context.Context, req ChargeRequest) (transactionID string, err error)
	Refund(ctx context.Context, req RefundRequest) error
	Lookup(ctx context.Context, req LookupRequest) (transactionID string, found bool, err error)
}

type CouponProvider interface {
	Quote(ctx context.Context, code string, amountCents int) (discountCents int, err error)
	Redeem(ctx context.Context, code string, orderNo int64, amountCents int) error
}

type Notifier interface {
	Notify(ctx context.Context, e domain.Event) error
}
