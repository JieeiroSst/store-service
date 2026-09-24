package port

import (
	"context"
	"io"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

type InvoiceStatus string

const (
	InvoiceIssued    InvoiceStatus = "issued"
	InvoicePaid      InvoiceStatus = "paid"
	InvoiceOverdue   InvoiceStatus = "overdue"
	InvoiceCancelled InvoiceStatus = "cancelled"
)

type InvoiceState struct {
	Amount  float64
	DueDate *time.Time
	Status  InvoiceStatus
}

type InvoiceGateway interface {
	Enabled() bool
	EnsureAccount(ctx context.Context, p *model.Patient, existing *model.BillingAccount) (*model.BillingAccount, error)
	CreateInvoice(ctx context.Context, acct *model.BillingAccount, state InvoiceState) (invoiceID int32, err error)
	SyncInvoice(ctx context.Context, invoiceID int32, state InvoiceState) error
	RecordPayment(ctx context.Context, invoiceID int32, amount float64, method string, at time.Time) error
}

type BillingAccountRepository interface {
	Get(ctx context.Context, patientID int32) (*model.BillingAccount, error)
	Save(ctx context.Context, a *model.BillingAccount) error
	WithLock(ctx context.Context, patientID int32, fn func(ctx context.Context) error) error
}

type ChargeInput struct {
	PatientID   int32
	UserID      string
	Amount      float64
	Reference   string
	Description string
}

type WalletGateway interface {
	Enabled() bool
	Charge(ctx context.Context, in ChargeInput) (transferID string, err error)
	Refund(ctx context.Context, transferID, reason string) error
}

type Notification struct {
	PatientID      int32
	Email          string
	Title, Message string
}

type Notifier interface {
	Notify(ctx context.Context, n Notification) error
}

type TokenValidator interface {
	Validate(ctx context.Context, token string) (userID string, err error)
}

type IdentityGateway interface {
	Enabled() bool
	SubmitCitizenCard(ctx context.Context, userID string, front, back []byte) error
	SubmitFaceScan(ctx context.Context, userID string, frames [][]byte) error
	Verify(ctx context.Context, userID string) (*model.IdentityStatus, error)
	Status(ctx context.Context, userID string) (*model.IdentityStatus, error)
}

type DocumentUpload struct {
	FileName string
	Body     io.Reader
}

type DocumentDownload struct {
	Document *model.Document
	Body     io.ReadCloser
}

type DocumentGateway interface {
	Enabled() bool
	Upload(ctx context.Context, patientID int32, in DocumentUpload) (*model.Document, error)
	List(ctx context.Context, patientID int32, limit, offset int) ([]model.Document, int64, error)
	Open(ctx context.Context, patientID int32, docID string) (*DocumentDownload, error)
	Delete(ctx context.Context, patientID int32, docID string) error
}

type bearerKey struct{}

func WithBearer(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, bearerKey{}, token)
}

func BearerFrom(ctx context.Context) string {
	t, _ := ctx.Value(bearerKey{}).(string)
	return t
}
