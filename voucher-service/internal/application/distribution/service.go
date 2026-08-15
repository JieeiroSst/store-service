package distribution

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Service struct {
	jobs        JobRepository
	claims      ClaimRepository
	bulkIssuer  VoucherBulkIssuer
	budgetCheck BudgetChecker
	clock       shared.Clock
}

func NewService(jobs JobRepository, claims ClaimRepository, bulkIssuer VoucherBulkIssuer, budgetCheck BudgetChecker, clock shared.Clock) DistributionService {
	return &Service{jobs: jobs, claims: claims, bulkIssuer: bulkIssuer, budgetCheck: budgetCheck, clock: clock}
}

func (s *Service) CreateJob(ctx context.Context, in CreateJobInput) (*Job, error) {
	recipientCount := len(in.Recipients)
	totalCost := shared.NewMoney(in.Denomination.Amount*int64(recipientCount), in.Denomination.Currency)

	if err := s.budgetCheck.CheckBudget(ctx, in.CorporateID, totalCost); err != nil {
		return nil, err
	}

	job := &Job{
		CorporateID:     in.CorporateID,
		Status:          StatusInProgress,
		TotalRecipients: recipientCount,
		CreatedAt:       s.clock.Now(),
	}
	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, err
	}

	voucherIDs, err := s.bulkIssuer.IssueBulk(ctx, in.MerchantID, in.ProductSKU, in.Denomination, recipientCount, in.CorporateID)
	if err != nil {
		job.Status = StatusFailed
		_ = s.jobs.Save(ctx, job)
		return nil, err
	}

	claimsBatch := make([]*Claim, 0, recipientCount)
	expiresAt := s.clock.Now().AddDate(0, 0, 30)
	for i, recipient := range in.Recipients {
		token, err := randomClaimToken()
		if err != nil {
			return nil, err
		}
		vid := voucherIDs[i]
		claimsBatch = append(claimsBatch, &Claim{
			DistributionJobID:   job.ID,
			VoucherID:           &vid,
			RecipientIdentifier: recipient,
			ClaimToken:          token,
			ClaimTokenExpiresAt: expiresAt,
			Status:              ClaimStatusSent,
		})
	}
	if err := s.claims.CreateBatch(ctx, claimsBatch); err != nil {
		return nil, err
	}

	job.Status = StatusCompleted
	if err := s.jobs.Save(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Service) ClaimVoucher(ctx context.Context, claimToken string) (shared.VoucherID, error) {
	claim, err := s.claims.FindByToken(ctx, claimToken)
	if err != nil {
		return shared.VoucherID{}, err
	}
	if claim.Status == ClaimStatusClaimed {
		return shared.VoucherID{}, errAlreadyClaimed
	}
	if s.clock.Now().After(claim.ClaimTokenExpiresAt) {
		claim.Status = ClaimStatusExpired
		_ = s.claims.Save(ctx, claim)
		return shared.VoucherID{}, errClaimExpired
	}

	now := s.clock.Now()
	claim.Status = ClaimStatusClaimed
	claim.ClaimedAt = &now
	if err := s.claims.Save(ctx, claim); err != nil {
		return shared.VoucherID{}, err
	}
	if claim.VoucherID == nil {
		return shared.VoucherID{}, errNoVoucherForClaim
	}
	return *claim.VoucherID, nil
}

func (s *Service) GetJob(ctx context.Context, id string) (*Job, error) {
	return s.jobs.FindByID(ctx, id)
}

func randomClaimToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
