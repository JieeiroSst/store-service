package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
)

type ListQuery struct {
	Offset  int
	Limit   int
	Equals  map[string]any
	Search  string
	Exclude map[string][]any
}

type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uint) (*T, error)
	List(ctx context.Context, q ListQuery) ([]T, int64, error)
	Update(ctx context.Context, id uint, entity *T) error
	Delete(ctx context.Context, id uint) error
}

type TxRunner interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Reporter interface {
	Summary(ctx context.Context) (*model.Summary, error)
}

type ContractExpirer interface {
	ExpireOne(ctx context.Context, id uint, now time.Time) (bool, error)
	ExpireDue(ctx context.Context, now time.Time) (int64, error)
}

type SignatureEvidence struct {
	Algorithm      string
	Signature      string
	CertificatePEM string
}

type TimestampEvidence struct {
	Token     string
	Time      time.Time
	Authority string
}

type VerifiedSignature struct {
	Format      string
	Revocation  string
	Timestamp   *TimestampEvidence
	Algorithm   string
	Signature   string
	Certificate string
	Subject     string
	Serial      string
	Fingerprint string
	PayloadHash string
}

type SignatureVerifier interface {
	Verify(ctx context.Context, payload []byte, ev SignatureEvidence, at time.Time) (*VerifiedSignature, error)
}

type PDFSignatureVerifier interface {
	VerifyPDF(ctx context.Context, doc []byte, at time.Time) (*VerifiedSignature, error)
}

type Timestamper interface {
	Stamp(ctx context.Context, data []byte) (*TimestampEvidence, error)
}

type ContractRenderer interface {
	Render(doc RenewalDocument) ([]byte, error)
}

type RenewalDocument struct {
	Contract    *model.Contract
	Account     *model.Account // "Bên B"
	SignedBy    string
	EndDate     time.Time
	SigningTime time.Time
}

type ContractOrchestrator interface {
	Sync(ctx context.Context, contractID uint) error
}

type CounterSigner interface {
	Enabled() bool
	CounterSign(ctx context.Context, signed []byte, at time.Time) ([]byte, *VerifiedSignature, error)
}

type ContractFileRepository interface {
	Repository[model.ContractFile]
	SetLatest(ctx context.Context, id uint, latest bool) (bool, error)
	Lineage(ctx context.Context, root uint) ([]model.ContractFile, error)
}

type ScanResult struct {
	Clean     bool
	Signature string
	Engine    string
}

type Scanner interface {
	Enabled() bool
	Scan(ctx context.Context, name string, data []byte) (ScanResult, error)
}

type DocumentStore interface {
	Enabled() bool
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type Notifier interface {
	Notify(ctx context.Context, notification model.Notification) error
}
