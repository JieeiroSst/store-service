package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/referral"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type partnerRepo struct{ db *sqlx.DB }

func NewPartnerRepository(db *sqlx.DB) referral.PartnerRepository {
	return &partnerRepo{db: db}
}

func (r *partnerRepo) Save(ctx context.Context, p *referral.Partner) error {
	const q = `
		INSERT INTO partners (
			id, user_id, full_name, email, phone, company, bio,
			referred_by_partner_id, network_depth, status, tier, commission,
			total_referrals, hired_referrals, conversion_rate,
			total_earned, pending_payout, created_at, updated_at
		) VALUES (
			:id, :user_id, :full_name, :email, :phone, :company, :bio,
			:referred_by_partner_id, :network_depth, :status, :tier, :commission,
			:total_referrals, :hired_referrals, :conversion_rate,
			:total_earned, :pending_payout, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, q, partnerToRow(p))
	return err
}

func (r *partnerRepo) Update(ctx context.Context, p *referral.Partner) error {
	const q = `
		UPDATE partners SET
			full_name = :full_name, phone = :phone, company = :company, bio = :bio,
			status = :status, tier = :tier, commission = :commission,
			total_referrals = :total_referrals, hired_referrals = :hired_referrals,
			conversion_rate = :conversion_rate,
			total_earned = :total_earned, pending_payout = :pending_payout,
			updated_at = NOW()
		WHERE id = :id AND deleted_at IS NULL`
	_, err := r.db.NamedExecContext(ctx, q, partnerToRow(p))
	return err
}

func (r *partnerRepo) FindByID(ctx context.Context, id uuid.UUID) (*referral.Partner, error) {
	var row partnerRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM partners WHERE id = $1 AND deleted_at IS NULL`, id,
	); err != nil {
		return nil, fmt.Errorf("partnerRepo.FindByID: %w", err)
	}
	return partnerFromRow(&row)
}

func (r *partnerRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*referral.Partner, error) {
	var row partnerRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM partners WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	); err != nil {
		return nil, err
	}
	return partnerFromRow(&row)
}

// FindNetwork returns all direct/indirect downline partners up to `depth` levels
func (r *partnerRepo) FindNetwork(ctx context.Context, partnerID uuid.UUID, depth int) ([]*referral.Partner, error) {
	const q = `
		WITH RECURSIVE network AS (
			SELECT * FROM partners WHERE referred_by_partner_id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT p.* FROM partners p
				INNER JOIN network n ON p.referred_by_partner_id = n.id
			WHERE p.deleted_at IS NULL AND n.network_depth < $2
		)
		SELECT * FROM network`
	var rows []partnerRow
	if err := r.db.SelectContext(ctx, &rows, q, partnerID, depth); err != nil {
		return nil, err
	}
	result := make([]*referral.Partner, 0, len(rows))
	for i := range rows {
		p, err := partnerFromRow(&rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *partnerRepo) FindTopPerformers(ctx context.Context, limit int) ([]*referral.Partner, error) {
	var rows []partnerRow
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM partners WHERE status = 'active' AND deleted_at IS NULL
		 ORDER BY hired_referrals DESC, conversion_rate DESC LIMIT $1`, limit,
	); err != nil {
		return nil, err
	}
	result := make([]*referral.Partner, 0, len(rows))
	for i := range rows {
		p, err := partnerFromRow(&rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

type partnerRow struct {
	ID                  uuid.UUID  `db:"id"`
	UserID              uuid.UUID  `db:"user_id"`
	FullName            string     `db:"full_name"`
	Email               string     `db:"email"`
	Phone               string     `db:"phone"`
	Company             string     `db:"company"`
	Bio                 string     `db:"bio"`
	ReferredByPartnerID *uuid.UUID `db:"referred_by_partner_id"`
	NetworkDepth        int        `db:"network_depth"`
	Status              string     `db:"status"`
	Tier                string     `db:"tier"`
	CommissionJSON      []byte     `db:"commission"`
	TotalReferrals      int        `db:"total_referrals"`
	HiredReferrals      int        `db:"hired_referrals"`
	ConversionRate      float64    `db:"conversion_rate"`
	TotalEarnedJSON     []byte     `db:"total_earned"`
	PendingPayoutJSON   []byte     `db:"pending_payout"`
}

func partnerToRow(p *referral.Partner) map[string]interface{} {
	commissionJSON, _ := json.Marshal(p.Commission)
	totalEarnedJSON, _ := json.Marshal(p.TotalEarned)
	pendingPayoutJSON, _ := json.Marshal(p.PendingPayout)
	return map[string]interface{}{
		"id": p.ID, "user_id": p.UserID, "full_name": p.FullName,
		"email": p.Email, "phone": p.Phone, "company": p.Company, "bio": p.Bio,
		"referred_by_partner_id": p.ReferredByPartnerID,
		"network_depth":          p.NetworkDepth, "status": p.Status, "tier": p.Tier,
		"commission":      commissionJSON,
		"total_referrals": p.TotalReferrals, "hired_referrals": p.HiredReferrals,
		"conversion_rate": p.ConversionRate,
		"total_earned":    totalEarnedJSON, "pending_payout": pendingPayoutJSON,
		"created_at": p.CreatedAt, "updated_at": p.UpdatedAt,
	}
}

func partnerFromRow(row *partnerRow) (*referral.Partner, error) {
	p := &referral.Partner{}
	p.ID = row.ID
	p.UserID = row.UserID
	p.FullName = row.FullName
	p.Email = row.Email
	p.Phone = row.Phone
	p.Company = row.Company
	p.Bio = row.Bio
	p.ReferredByPartnerID = row.ReferredByPartnerID
	p.NetworkDepth = row.NetworkDepth
	p.Status = referral.PartnerStatus(row.Status)
	p.Tier = referral.PartnerTier(row.Tier)
	p.TotalReferrals = row.TotalReferrals
	p.HiredReferrals = row.HiredReferrals
	p.ConversionRate = row.ConversionRate
	_ = json.Unmarshal(row.CommissionJSON, &p.Commission)
	_ = json.Unmarshal(row.TotalEarnedJSON, &p.TotalEarned)
	_ = json.Unmarshal(row.PendingPayoutJSON, &p.PendingPayout)
	return p, nil
}

type referralRepo struct{ db *sqlx.DB }

func NewReferralRepository(db *sqlx.DB) referral.ReferralRepository {
	return &referralRepo{db: db}
}

func (r *referralRepo) Save(ctx context.Context, ref *referral.Referral) error {
	const q = `
		INSERT INTO referrals (id, partner_id, candidate_id, job_id, application_id,
			token, status, expires_at, commission_due, commission_paid_at,
			click_count, last_clicked_at, created_at, updated_at)
		VALUES (:id, :partner_id, :candidate_id, :job_id, :application_id,
			:token, :status, :expires_at, :commission_due, :commission_paid_at,
			:click_count, :last_clicked_at, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, referralToRow(ref))
	return err
}

func (r *referralRepo) Update(ctx context.Context, ref *referral.Referral) error {
	const q = `
		UPDATE referrals SET
			candidate_id = :candidate_id, application_id = :application_id,
			status = :status, commission_due = :commission_due,
			commission_paid_at = :commission_paid_at,
			click_count = :click_count, last_clicked_at = :last_clicked_at,
			updated_at = NOW()
		WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, q, referralToRow(ref))
	return err
}

func (r *referralRepo) FindByToken(ctx context.Context, token string) (*referral.Referral, error) {
	var row refRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM referrals WHERE token = $1`, token,
	); err != nil {
		return nil, err
	}
	return referralFromRow(&row)
}

func (r *referralRepo) FindByPartner(ctx context.Context, partnerID uuid.UUID) ([]*referral.Referral, error) {
	var rows []refRow
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM referrals WHERE partner_id = $1 ORDER BY created_at DESC`, partnerID,
	); err != nil {
		return nil, err
	}
	result := make([]*referral.Referral, 0, len(rows))
	for i := range rows {
		ref, err := referralFromRow(&rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, ref)
	}
	return result, nil
}

func (r *referralRepo) FindByApplication(ctx context.Context, applicationID uuid.UUID) ([]*referral.Referral, error) {
	var rows []refRow
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM referrals WHERE application_id = $1`, applicationID,
	); err != nil {
		return nil, err
	}
	result := make([]*referral.Referral, 0, len(rows))
	for i := range rows {
		ref, err := referralFromRow(&rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, ref)
	}
	return result, nil
}

type refRow struct {
	ID                uuid.UUID  `db:"id"`
	PartnerID         uuid.UUID  `db:"partner_id"`
	CandidateID       *uuid.UUID `db:"candidate_id"`
	JobID             *uuid.UUID `db:"job_id"`
	ApplicationID     *uuid.UUID `db:"application_id"`
	Token             string     `db:"token"`
	Status            string     `db:"status"`
	ExpiresAt         time.Time  `db:"expires_at"`
	CommissionDueJSON []byte     `db:"commission_due"`
	CommissionPaidAt  *time.Time `db:"commission_paid_at"`
	ClickCount        int        `db:"click_count"`
	LastClickedAt     *time.Time `db:"last_clicked_at"`
	CreatedAt         time.Time  `db:"created_at"`
}

func referralToRow(ref *referral.Referral) map[string]interface{} {
	commJSON, _ := json.Marshal(ref.CommissionDue)
	return map[string]interface{}{
		"id": ref.ID, "partner_id": ref.PartnerID,
		"candidate_id": ref.CandidateID, "job_id": ref.JobID, "application_id": ref.ApplicationID,
		"token": ref.Token, "status": ref.Status, "expires_at": ref.ExpiresAt,
		"commission_due": commJSON, "commission_paid_at": ref.CommissionPaidAt,
		"click_count": ref.ClickCount, "last_clicked_at": ref.LastClickedAt,
		"created_at": ref.CreatedAt, "updated_at": ref.CreatedAt,
	}
}

func referralFromRow(row *refRow) (*referral.Referral, error) {
	ref := &referral.Referral{}
	ref.ID = row.ID
	ref.PartnerID = row.PartnerID
	ref.CandidateID = row.CandidateID
	ref.JobID = row.JobID
	ref.ApplicationID = row.ApplicationID
	ref.Token = row.Token
	ref.Status = referral.ReferralStatus(row.Status)
	ref.ExpiresAt = row.ExpiresAt
	ref.CommissionPaidAt = row.CommissionPaidAt
	ref.ClickCount = row.ClickCount
	ref.LastClickedAt = row.LastClickedAt
	ref.CreatedAt = row.CreatedAt
	_ = json.Unmarshal(row.CommissionDueJSON, &ref.CommissionDue)
	return ref, nil
}

type payoutRepo struct{ db *sqlx.DB }

func NewPayoutRepository(db *sqlx.DB) referral.PayoutRepository {
	return &payoutRepo{db: db}
}

func (r *payoutRepo) Save(ctx context.Context, p *referral.Payout) error {
	amountJSON, _ := json.Marshal(p.Amount)
	referralIDsJSON, _ := json.Marshal(p.ReferralIDs)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payouts (id, partner_id, referral_ids, amount, status, note, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
		p.ID, p.PartnerID, referralIDsJSON, amountJSON, p.Status, p.Note, p.CreatedAt,
	)
	return err
}

func (r *payoutRepo) Update(ctx context.Context, p *referral.Payout) error {
	amountJSON, _ := json.Marshal(p.Amount)
	_, err := r.db.ExecContext(ctx,
		`UPDATE payouts SET status = $1, amount = $2, processed_at = $3, note = $4, updated_at = NOW() WHERE id = $5`,
		p.Status, amountJSON, p.ProcessedAt, p.Note, p.ID,
	)
	return err
}

func (r *payoutRepo) FindPendingByPartner(ctx context.Context, partnerID uuid.UUID) ([]*referral.Payout, error) {
	type payRow struct {
		ID          uuid.UUID  `db:"id"`
		PartnerID   uuid.UUID  `db:"partner_id"`
		ReferralIDs []byte     `db:"referral_ids"`
		AmountJSON  []byte     `db:"amount"`
		Status      string     `db:"status"`
		Note        string     `db:"note"`
		ProcessedAt *time.Time `db:"processed_at"`
		CreatedAt   time.Time  `db:"created_at"`
	}
	var rows []payRow
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM payouts WHERE partner_id = $1 AND status = 'pending' ORDER BY created_at DESC`,
		partnerID,
	); err != nil {
		return nil, err
	}
	result := make([]*referral.Payout, 0, len(rows))
	for _, row := range rows {
		pay := &referral.Payout{}
		pay.ID = row.ID
		pay.PartnerID = row.PartnerID
		pay.Status = referral.PayoutStatus(row.Status)
		pay.Note = row.Note
		pay.ProcessedAt = row.ProcessedAt
		pay.CreatedAt = row.CreatedAt
		_ = json.Unmarshal(row.AmountJSON, &pay.Amount)
		_ = json.Unmarshal(row.ReferralIDs, &pay.ReferralIDs)
		result = append(result, pay)
	}
	return result, nil
}
