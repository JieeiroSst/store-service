package inbound

import (
	"context"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type LeaseCommand struct {
	HomestayID int64
	Model      domain.RentalModel // week, month or year
	Start      time.Time
	Periods    int
	BillingDay int
	Note       string
	RequestID  string
}

type LeaseUseCase interface {
	Start(ctx context.Context, actor Principal, cmd LeaseCommand) (domain.Lease, error)
	Get(ctx context.Context, actor Principal, id int64) (domain.Lease, error)
	List(ctx context.Context, actor Principal, userID int64, asHost bool, limit, offset int) ([]domain.Lease, error)
	PayInvoice(ctx context.Context, actor Principal, leaseID int64, periodNo int, cmd PayCommand) (domain.Lease, error)
	Cancel(ctx context.Context, actor Principal, id int64) (domain.Lease, error)
	Terminate(ctx context.Context, actor Principal, id int64) (domain.Lease, error)
	Invoices(ctx context.Context, actor Principal, f InvoiceQuery) ([]domain.Invoice, error)
	ReleaseExpired(ctx context.Context) (int, error)
}

type InvoiceQuery struct {
	Status    domain.InvoiceStatus
	DueBefore time.Time
	Limit     int
	Offset    int
}
