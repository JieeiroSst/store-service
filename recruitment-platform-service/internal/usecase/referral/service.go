package referralusecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/referral"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type service struct {
	partnerRepo  referral.PartnerRepository
	referralRepo referral.ReferralRepository
	payoutRepo   referral.PayoutRepository
	notifySvc    port.NotificationService
	eventBus     port.EventBus
	logger       *zap.Logger
}

func NewService(
	partnerRepo referral.PartnerRepository,
	referralRepo referral.ReferralRepository,
	payoutRepo referral.PayoutRepository,
	notifySvc port.NotificationService,
	eventBus port.EventBus,
	logger *zap.Logger,
) port.ReferralService {
	return &service{
		partnerRepo:  partnerRepo,
		referralRepo: referralRepo,
		payoutRepo:   payoutRepo,
		notifySvc:    notifySvc,
		eventBus:     eventBus,
		logger:       logger,
	}
}

func (s *service) RegisterPartner(ctx context.Context, cmd port.RegisterPartnerCommand) (*referral.Partner, error) {
	p, err := referral.NewPartner(cmd.UserID, cmd.FullName, cmd.Email)
	if err != nil {
		return nil, err
	}
	p.Phone = cmd.Phone
	p.Company = cmd.Company

	if cmd.ReferredByToken != "" {
		uplineRef, err := s.referralRepo.FindByToken(ctx, cmd.ReferredByToken)
		if err == nil && uplineRef != nil && !uplineRef.IsExpired() {
			p.ReferredByPartnerID = &uplineRef.PartnerID

			upline, err := s.partnerRepo.FindByID(ctx, uplineRef.PartnerID)
			if err == nil {
				p.NetworkDepth = upline.NetworkDepth + 1
			}
		}
	}

	pct := 5.0
	p.Commission = referral.CommissionConfig{
		Type:          "percentage",
		Percentage:    &pct,
		PayoutTrigger: "probation_passed",
		ProbationDays: 60,
	}

	if err := s.partnerRepo.Save(ctx, p); err != nil {
		return nil, err
	}

	// Welcome notification
	_ = s.notifySvc.Send(ctx, port.NotificationPayload{
		RecipientID: p.UserID,
		Channel:     "email",
		TemplateID:  "partner_welcome",
		Data:        map[string]any{"full_name": p.FullName, "tier": p.Tier},
	})

	return p, nil
}

func (s *service) GenerateReferralLink(ctx context.Context, partnerID uuid.UUID, jobID *uuid.UUID) (*referral.Referral, error) {
	if _, err := s.partnerRepo.FindByID(ctx, partnerID); err != nil {
		return nil, errors.New("referral: partner not found")
	}

	token, err := generateToken(12)
	if err != nil {
		return nil, err
	}

	ref := referral.NewReferral(partnerID, jobID, token, 30) // 30-day TTL
	if err := s.referralRepo.Save(ctx, ref); err != nil {
		return nil, err
	}
	return ref, nil
}

func (s *service) TrackReferralClick(ctx context.Context, token string) error {
	ref, err := s.referralRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if ref.IsExpired() {
		return errors.New("referral: link expired")
	}
	ref.RecordClick()
	return s.referralRepo.Update(ctx, ref)
}

func (s *service) TrackApplication(ctx context.Context, token string, candidateID, applicationID uuid.UUID) error {
	ref, err := s.referralRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if ref.IsExpired() {
		return errors.New("referral: link expired")
	}
	ref.CandidateID = &candidateID
	ref.ApplicationID = &applicationID
	ref.Status = referral.ReferralStatusApplied

	// Increment partner referral count
	partner, err := s.partnerRepo.FindByID(ctx, ref.PartnerID)
	if err != nil {
		return err
	}
	partner.TotalReferrals++
	if err := s.partnerRepo.Update(ctx, partner); err != nil {
		return err
	}

	return s.referralRepo.Update(ctx, ref)
}

func (s *service) TrackHire(ctx context.Context, applicationID uuid.UUID) error {
	refs, err := s.referralRepo.FindByApplication(ctx, applicationID)
	if err != nil || len(refs) == 0 {
		s.logger.Info("no referral found for application", zap.String("application_id", applicationID.String()))
		return nil
	}
	ref := refs[0]
	ref.Status = referral.ReferralStatusHired

	partner, err := s.partnerRepo.FindByID(ctx, ref.PartnerID)
	if err != nil {
		return err
	}

	commission := s.computeCommission(partner)
	ref.CommissionDue = &commission

	partner.RecordHire(commission)

	if err := s.partnerRepo.Update(ctx, partner); err != nil {
		return err
	}
	if err := s.referralRepo.Update(ctx, ref); err != nil {
		return err
	}

	_ = s.notifySvc.Send(ctx, port.NotificationPayload{
		RecipientID: partner.UserID,
		Channel:     "email",
		TemplateID:  "referral_hire_confirmed",
		Data: map[string]any{
			"commission_amount": commission.Amount,
			"currency":          commission.Currency,
			"payout_trigger":    partner.Commission.PayoutTrigger,
		},
	})

	return nil
}

func (s *service) GetPartnerStats(ctx context.Context, partnerID uuid.UUID) (*port.PartnerStats, error) {
	partner, err := s.partnerRepo.FindByID(ctx, partnerID)
	if err != nil {
		return nil, err
	}
	return &port.PartnerStats{
		Partner:        partner,
		TotalReferrals: partner.TotalReferrals,
		HiredCount:     partner.HiredReferrals,
		ConversionRate: partner.ConversionRate,
		PendingPayout:  partner.PendingPayout,
		TotalEarned:    partner.TotalEarned,
	}, nil
}

func (s *service) GetLeaderboard(ctx context.Context, limit int) ([]*port.PartnerStats, error) {
	partners, err := s.partnerRepo.FindTopPerformers(ctx, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*port.PartnerStats, 0, len(partners))
	for i, p := range partners {
		result = append(result, &port.PartnerStats{
			Partner:        p,
			TotalReferrals: p.TotalReferrals,
			HiredCount:     p.HiredReferrals,
			ConversionRate: p.ConversionRate,
			PendingPayout:  p.PendingPayout,
			TotalEarned:    p.TotalEarned,
			Rank:           i + 1,
		})
	}
	return result, nil
}

func (s *service) RequestPayout(ctx context.Context, partnerID uuid.UUID) (*referral.Payout, error) {
	partner, err := s.partnerRepo.FindByID(ctx, partnerID)
	if err != nil {
		return nil, err
	}
	if partner.PendingPayout.Amount <= 0 {
		return nil, errors.New("referral: no pending payout available")
	}

	payout := &referral.Payout{
		BaseEntity: shared.NewBaseEntity(),
		PartnerID:  partnerID,
		Amount:     partner.PendingPayout,
		Status:     referral.PayoutStatusPending,
	}
	if err := s.payoutRepo.Save(ctx, payout); err != nil {
		return nil, err
	}

	partner.PendingPayout = shared.Money{Amount: 0, Currency: partner.PendingPayout.Currency}
	if err := s.partnerRepo.Update(ctx, partner); err != nil {
		return nil, err
	}

	_ = s.notifySvc.Send(ctx, port.NotificationPayload{
		RecipientID: partner.UserID,
		Channel:     "email",
		TemplateID:  "payout_requested",
		Data:        map[string]any{"amount": payout.Amount.Amount, "currency": payout.Amount.Currency},
	})

	return payout, nil
}

func (s *service) computeCommission(p *referral.Partner) shared.Money {
	currency := "VND"
	if p.Commission.FixedAmount != nil {
		return *p.Commission.FixedAmount
	}
	// Fallback fixed bonus by tier when percentage isn't calculable without salary context
	var amount int64
	switch p.Tier {
	case referral.TierPlatinum:
		amount = 20_000_000
	case referral.TierGold:
		amount = 15_000_000
	case referral.TierSilver:
		amount = 10_000_000
	default:
		amount = 5_000_000
	}
	return shared.Money{Amount: amount, Currency: currency}
}

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
