package referral

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
)

type PartnerStatus string

const (
	PartnerStatusActive    PartnerStatus = "active"
	PartnerStatusInactive  PartnerStatus = "inactive"
	PartnerStatusSuspended PartnerStatus = "suspended"
)

type PartnerTier string

const (
	TierBronze   PartnerTier = "bronze"
	TierSilver   PartnerTier = "silver"
	TierGold     PartnerTier = "gold"
	TierPlatinum PartnerTier = "platinum"
)

type CommissionConfig struct {
	Type          string        `json:"type"` // "fixed" | "percentage"
	FixedAmount   *shared.Money `json:"fixed_amount,omitempty"`
	Percentage    *float64      `json:"percentage,omitempty"`
	PayoutTrigger string        `json:"payout_trigger"` // "hired" | "probation_passed"
	ProbationDays int           `json:"probation_days"`
}

type Partner struct {
	shared.BaseEntity

	UserID   uuid.UUID `db:"user_id"   json:"user_id"`
	FullName string    `db:"full_name" json:"full_name"`
	Email    string    `db:"email"     json:"email"`
	Phone    string    `db:"phone"     json:"phone"`
	Company  string    `db:"company"   json:"company"`
	Bio      string    `db:"bio"       json:"bio"`

	ReferredByPartnerID *uuid.UUID `db:"referred_by_partner_id" json:"referred_by_partner_id,omitempty"`
	NetworkDepth        int        `db:"network_depth"          json:"network_depth"`

	Status PartnerStatus `db:"status" json:"status"`
	Tier   PartnerTier   `db:"tier"   json:"tier"`

	Commission CommissionConfig `db:"commission" json:"commission"`

	TotalReferrals int          `db:"total_referrals" json:"total_referrals"`
	HiredReferrals int          `db:"hired_referrals" json:"hired_referrals"`
	ConversionRate float64      `db:"conversion_rate" json:"conversion_rate"`
	TotalEarned    shared.Money `db:"total_earned"    json:"total_earned"`
	PendingPayout  shared.Money `db:"pending_payout"  json:"pending_payout"`

	events []shared.DomainEvent
}

func NewPartner(userID uuid.UUID, fullName, email string) (*Partner, error) {
	if fullName == "" || email == "" {
		return nil, errors.New("partner: name and email required")
	}
	return &Partner{
		BaseEntity: shared.NewBaseEntity(),
		UserID:     userID,
		FullName:   fullName,
		Email:      email,
		Status:     PartnerStatusActive,
		Tier:       TierBronze,
	}, nil
}

func (p *Partner) UpdateTier() {
	switch {
	case p.HiredReferrals >= 50:
		p.Tier = TierPlatinum
	case p.HiredReferrals >= 20:
		p.Tier = TierGold
	case p.HiredReferrals >= 5:
		p.Tier = TierSilver
	default:
		p.Tier = TierBronze
	}
}

func (p *Partner) RecordHire(commission shared.Money) {
	p.HiredReferrals++
	p.PendingPayout.Amount += commission.Amount
	p.TotalEarned.Amount += commission.Amount
	if p.TotalReferrals > 0 {
		p.ConversionRate = float64(p.HiredReferrals) / float64(p.TotalReferrals)
	}
	p.UpdateTier()
	p.record("PartnerHireRecorded", map[string]any{
		"partner_id": p.ID,
		"commission": commission,
		"new_tier":   p.Tier,
	})
}

func (p *Partner) DomainEvents() []shared.DomainEvent { return p.events }
func (p *Partner) ClearEvents()                       { p.events = nil }
func (p *Partner) record(t string, payload interface{}) {
	p.events = append(p.events, shared.NewDomainEvent(t, payload))
}

type ReferralStatus string

const (
	ReferralStatusPending  ReferralStatus = "pending"
	ReferralStatusApplied  ReferralStatus = "applied"
	ReferralStatusHired    ReferralStatus = "hired"
	ReferralStatusRejected ReferralStatus = "rejected"
	ReferralStatusExpired  ReferralStatus = "expired"
)

type Referral struct {
	shared.BaseEntity

	PartnerID     uuid.UUID  `db:"partner_id"     json:"partner_id"`
	CandidateID   *uuid.UUID `db:"candidate_id"   json:"candidate_id,omitempty"`
	JobID         *uuid.UUID `db:"job_id"         json:"job_id,omitempty"`
	ApplicationID *uuid.UUID `db:"application_id" json:"application_id,omitempty"`

	Token     string         `db:"token"      json:"token"`
	Status    ReferralStatus `db:"status"     json:"status"`
	ExpiresAt time.Time      `db:"expires_at" json:"expires_at"`

	CommissionDue    *shared.Money `db:"commission_due"     json:"commission_due,omitempty"`
	CommissionPaidAt *time.Time    `db:"commission_paid_at" json:"commission_paid_at,omitempty"`

	ClickCount    int        `db:"click_count"     json:"click_count"`
	LastClickedAt *time.Time `db:"last_clicked_at" json:"last_clicked_at,omitempty"`
}

func NewReferral(partnerID uuid.UUID, jobID *uuid.UUID, token string, ttlDays int) *Referral {
	return &Referral{
		BaseEntity: shared.NewBaseEntity(),
		PartnerID:  partnerID,
		JobID:      jobID,
		Token:      token,
		Status:     ReferralStatusPending,
		ExpiresAt:  time.Now().AddDate(0, 0, ttlDays),
	}
}

func (r *Referral) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r *Referral) RecordClick() {
	r.ClickCount++
	now := time.Now()
	r.LastClickedAt = &now
}

type PayoutStatus string

const (
	PayoutStatusPending   PayoutStatus = "pending"
	PayoutStatusApproved  PayoutStatus = "approved"
	PayoutStatusProcessed PayoutStatus = "processed"
	PayoutStatusFailed    PayoutStatus = "failed"
)

type Payout struct {
	shared.BaseEntity
	PartnerID   uuid.UUID    `db:"partner_id"   json:"partner_id"`
	ReferralIDs []uuid.UUID  `db:"referral_ids" json:"referral_ids"`
	Amount      shared.Money `db:"amount"       json:"amount"`
	Status      PayoutStatus `db:"status"       json:"status"`
	Note        string       `db:"note"         json:"note"`
	ProcessedAt *time.Time   `db:"processed_at" json:"processed_at,omitempty"`
}

type PartnerRepository interface {
	Save(ctx context.Context, p *Partner) error
	Update(ctx context.Context, p *Partner) error
	FindByID(ctx context.Context, id uuid.UUID) (*Partner, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*Partner, error)
	FindNetwork(ctx context.Context, partnerID uuid.UUID, depth int) ([]*Partner, error)
	FindTopPerformers(ctx context.Context, limit int) ([]*Partner, error)
}

type ReferralRepository interface {
	Save(ctx context.Context, r *Referral) error
	Update(ctx context.Context, r *Referral) error
	FindByToken(ctx context.Context, token string) (*Referral, error)
	FindByPartner(ctx context.Context, partnerID uuid.UUID) ([]*Referral, error)
	FindByApplication(ctx context.Context, applicationID uuid.UUID) ([]*Referral, error) // needed for hire tracking
}

type PayoutRepository interface {
	Save(ctx context.Context, p *Payout) error
	Update(ctx context.Context, p *Payout) error
	FindPendingByPartner(ctx context.Context, partnerID uuid.UUID) ([]*Payout, error)
}
