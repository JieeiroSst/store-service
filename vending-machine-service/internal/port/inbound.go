package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

type MachineService interface {
	Register(ctx context.Context, m *domain.Machine) (*domain.Machine, error)
	Get(ctx context.Context, id string) (*domain.Machine, error)
	List(ctx context.Context, status string, page domain.PageRequest) (domain.Page[domain.Machine], error)
	Update(ctx context.Context, id string, patch domain.MachinePatch) (*domain.Machine, error)
	ChangeStatus(ctx context.Context, id string, status domain.MachineStatus) (*domain.Machine, error)
	RecordMaintenance(ctx context.Context, l *domain.MaintenanceLog) (*domain.MaintenanceLog, error)
	ListMaintenance(ctx context.Context, machineID string) ([]domain.MaintenanceLog, error)
	ListEvents(ctx context.Context, machineID string, limit int) ([]domain.Event, error)
	SalesReport(ctx context.Context, machineID string, from, to time.Time) (*domain.SalesReport, error)
}

type CatalogService interface {
	CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error)
	ListCategories(ctx context.Context, page domain.PageRequest) (domain.Page[domain.Category], error)
	CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, id string, patch domain.ProductPatch) (*domain.Product, error)
	ListProducts(ctx context.Context, categoryID string, page domain.PageRequest) (domain.Page[domain.Product], error)
}

type SlotInput struct {
	ProductID    string
	Quantity     int
	MaxCapacity  int
	LowThreshold int
}

type InventoryService interface {
	AssignSlot(ctx context.Context, machineID, slot string, in SlotInput) (*domain.Inventory, error)
	ListByMachine(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error)
	ListLow(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error)
	Restock(ctx context.Context, inventoryID string, units int) (*domain.Inventory, error)
}

type CheckoutInput struct {
	Method     domain.PaymentMethod
	WalletID   string
	CouponCode string
}

type VendingService interface {
	StartSession(ctx context.Context, machineID string) (*domain.Session, error)
	GetSession(ctx context.Context, id string) (*domain.Session, error)
	EndSession(ctx context.Context, id string) (*domain.Session, error)

	Reserve(ctx context.Context, sessionID, slot string) (*domain.Reservation, error)
	CancelReservation(ctx context.Context, id string) (*domain.Reservation, error)
	Checkout(ctx context.Context, reservationID string, in CheckoutInput) (*domain.Checkout, error)

	GetPayment(ctx context.Context, id string) (*domain.Checkout, error)
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	ListSessionOrders(ctx context.Context, sessionID string) ([]domain.Order, error)
	ReportDispense(ctx context.Context, orderID string, dispensed bool) (*domain.Order, error)
	RefundOrder(ctx context.Context, orderID, reason string) (*domain.Order, error)

	ExpireStale(ctx context.Context) (reservations, sessions int, err error)
	ReconcilePayments(ctx context.Context) (settled int, err error)
}

type AlertService interface {
	DispatchPending(ctx context.Context) (int, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
