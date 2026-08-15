package distribution

import (
	"context"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type ClaimStatus string

const (
	ClaimStatusPending ClaimStatus = "pending"
	ClaimStatusSent    ClaimStatus = "sent"
	ClaimStatusClaimed ClaimStatus = "claimed"
	ClaimStatusExpired ClaimStatus = "expired"
)

type Job struct {
	ID              string             `json:"id"`
	CorporateID     shared.CorporateID `json:"corporate_id"`
	OrderID         *shared.OrderID    `json:"order_id,omitempty"`
	Status          Status             `json:"status"`
	TotalRecipients int                `json:"total_recipients"`
	CreatedAt       time.Time          `json:"created_at"`
}

type Claim struct {
	ID                   string
	DistributionJobID    string
	VoucherID            *shared.VoucherID
	RecipientIdentifier  string
	ClaimToken           string
	ClaimTokenExpiresAt  time.Time
	Status               ClaimStatus
	ClaimedAt            *time.Time
}

// CreateJobInput starts a B2B bulk distribution: one voucher per recipient,
// each claimable exactly once via a high-entropy claim_token.
type CreateJobInput struct {
	CorporateID shared.CorporateID
	MerchantID  shared.MerchantID
	ProductSKU  string
	Denomination shared.Money
	Recipients  []string // opaque identifiers (email/phone/employee id)
}

type DistributionService interface {
	CreateJob(ctx context.Context, in CreateJobInput) (*Job, error)
	ClaimVoucher(ctx context.Context, claimToken string) (shared.VoucherID, error)
	GetJob(ctx context.Context, id string) (*Job, error)
}

type JobRepository interface {
	Create(ctx context.Context, j *Job) error
	FindByID(ctx context.Context, id string) (*Job, error)
	Save(ctx context.Context, j *Job) error
}

type ClaimRepository interface {
	CreateBatch(ctx context.Context, claims []*Claim) error
	FindByToken(ctx context.Context, token string) (*Claim, error)
	Save(ctx context.Context, c *Claim) error
}

type VoucherBulkIssuer interface {
	IssueBulk(ctx context.Context, merchantID shared.MerchantID, productSKU string, denomination shared.Money, quantity int, corporateID shared.CorporateID) ([]shared.VoucherID, error)
}

type BudgetChecker interface {
	CheckBudget(ctx context.Context, corporateID shared.CorporateID, proposedSpend shared.Money) error
}
