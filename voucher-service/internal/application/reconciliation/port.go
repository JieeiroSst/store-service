package reconciliation

import (
	"context"
	"time"
)

type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
)

type Discrepancy struct {
	Reference string
	Reason    string
}

type Run struct {
	ID                string
	RunType           string
	Status            RunStatus
	DiscrepancyCount  int
	Discrepancies     []Discrepancy
	StartedAt         *time.Time
	FinishedAt        *time.Time
}

type ReconciliationService interface {
	RunPaymentReconciliation(ctx context.Context) (*Run, error)
	GetRun(ctx context.Context, id string) (*Run, error)
}

type RunRepository interface {
	Create(ctx context.Context, r *Run) error
	Save(ctx context.Context, r *Run) error
	FindByID(ctx context.Context, id string) (*Run, error)
}

type PaymentRecord struct {
	PaymentID      string
	OrderID        string
	Amount         int64
	Currency       string
	ProviderTxnRef string
}

type PaymentRecordSource interface {
	ListSettledSince(ctx context.Context, since time.Time) ([]PaymentRecord, error)
}
