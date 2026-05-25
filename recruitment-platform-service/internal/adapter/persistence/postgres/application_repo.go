package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainapp "github.com/JIeeiroSst/recruitment-platform-service/internal/domain/application"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type applicationRepo struct {
	db *sqlx.DB
}

func NewApplicationRepository(db *sqlx.DB) domainapp.Repository {
	return &applicationRepo{db: db}
}

func (r *applicationRepo) Save(ctx context.Context, a *domainapp.Application) error {
	const q = `
		INSERT INTO applications (
			id, job_id, candidate_id, recruiter_id, status, current_stage_id,
			rejection_reason, withdraw_reason, match_score, priority,
			referred_by_partner_id, days_in_stage, last_moved_at,
			created_at, updated_at
		) VALUES (
			:id, :job_id, :candidate_id, :recruiter_id, :status, :current_stage_id,
			:rejection_reason, :withdraw_reason, :match_score, :priority,
			:referred_by_partner_id, :days_in_stage, :last_moved_at,
			:created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, q, appToRow(a))
	return err
}

func (r *applicationRepo) Update(ctx context.Context, a *domainapp.Application) error {
	const q = `
		UPDATE applications SET
			status = :status, current_stage_id = :current_stage_id,
			rejection_reason = :rejection_reason, withdraw_reason = :withdraw_reason,
			match_score = :match_score, priority = :priority,
			days_in_stage = :days_in_stage, last_moved_at = :last_moved_at,
			updated_at = NOW()
		WHERE id = :id AND deleted_at IS NULL`
	_, err := r.db.NamedExecContext(ctx, q, appToRow(a))
	return err
}

func (r *applicationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domainapp.Application, error) {
	var row appRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM applications WHERE id = $1 AND deleted_at IS NULL`, id,
	); err != nil {
		return nil, fmt.Errorf("applicationRepo.FindByID: %w", err)
	}
	a, err := appFromRow(&row)
	if err != nil {
		return nil, err
	}

	// Load interviews
	interviews, err := r.findInterviews(ctx, id)
	if err == nil {
		a.Interviews = interviews
	}

	// Load offer
	offer, err := r.findOffer(ctx, id)
	if err == nil && offer != nil {
		a.Offer = offer
	}

	return a, nil
}

func (r *applicationRepo) FindAll(ctx context.Context, f domainapp.Filter) (shared.PaginatedResult[*domainapp.Application], error) {
	args := []interface{}{}
	conds := []string{"deleted_at IS NULL"}
	idx := 1

	if f.JobID != nil {
		conds = append(conds, fmt.Sprintf("job_id = $%d", idx))
		args = append(args, *f.JobID)
		idx++
	}
	if f.CandidateID != nil {
		conds = append(conds, fmt.Sprintf("candidate_id = $%d", idx))
		args = append(args, *f.CandidateID)
		idx++
	}
	if f.RecruiterID != nil {
		conds = append(conds, fmt.Sprintf("recruiter_id = $%d", idx))
		args = append(args, *f.RecruiterID)
		idx++
	}
	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, *f.Status)
		idx++
	}

	where := strings.Join(conds, " AND ")
	var total int64
	if err := r.db.GetContext(ctx, &total,
		fmt.Sprintf(`SELECT COUNT(*) FROM applications WHERE %s`, where), args...,
	); err != nil {
		return shared.PaginatedResult[*domainapp.Application]{}, err
	}

	offset := f.Offset()
	rows := []appRow{}
	dataQ := fmt.Sprintf(
		`SELECT * FROM applications WHERE %s ORDER BY priority DESC, created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.Limit, offset)
	if err := r.db.SelectContext(ctx, &rows, dataQ, args...); err != nil {
		return shared.PaginatedResult[*domainapp.Application]{}, err
	}

	items := make([]*domainapp.Application, 0, len(rows))
	for i := range rows {
		a, err := appFromRow(&rows[i])
		if err != nil {
			return shared.PaginatedResult[*domainapp.Application]{}, err
		}
		items = append(items, a)
	}

	totalPages := int(total) / f.Limit
	if int(total)%f.Limit > 0 {
		totalPages++
	}
	return shared.PaginatedResult[*domainapp.Application]{
		Items: items, Total: total,
		Page: f.Page, Limit: f.Limit, TotalPages: totalPages,
	}, nil
}

func (r *applicationRepo) FindByJobAndCandidate(ctx context.Context, jobID, candidateID uuid.UUID) (*domainapp.Application, error) {
	var row appRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM applications WHERE job_id = $1 AND candidate_id = $2 AND deleted_at IS NULL`,
		jobID, candidateID,
	); err != nil {
		return nil, err
	}
	return appFromRow(&row)
}

func (r *applicationRepo) SaveInterview(ctx context.Context, appID uuid.UUID, iv domainapp.Interview) error {
	ivJSON, _ := json.Marshal(iv.Feedback)
	interviewerJSON, _ := json.Marshal(iv.InterviewerIDs)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO interviews (id, application_id, round, title, interviewer_ids, scheduled_at, duration_min, meeting_url, type, feedback)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		iv.ID, appID, iv.Round, iv.Title, interviewerJSON,
		iv.ScheduledAt, iv.DurationMin, iv.MeetingURL, iv.Type, ivJSON,
	)
	return err
}

func (r *applicationRepo) UpdateInterview(ctx context.Context, appID uuid.UUID, iv domainapp.Interview) error {
	ivJSON, _ := json.Marshal(iv.Feedback)
	_, err := r.db.ExecContext(ctx,
		`UPDATE interviews SET feedback = $1, updated_at = NOW() WHERE id = $2 AND application_id = $3`,
		ivJSON, iv.ID, appID,
	)
	return err
}

func (r *applicationRepo) findInterviews(ctx context.Context, appID uuid.UUID) ([]domainapp.Interview, error) {
	type interviewRow struct {
		ID             uuid.UUID `db:"id"`
		Round          int       `db:"round"`
		Title          string    `db:"title"`
		InterviewerIDs []byte    `db:"interviewer_ids"`
		ScheduledAt    time.Time `db:"scheduled_at"`
		DurationMin    int       `db:"duration_min"`
		MeetingURL     string    `db:"meeting_url"`
		Type           string    `db:"type"`
		FeedbackJSON   []byte    `db:"feedback"`
	}
	var rows []interviewRow
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM interviews WHERE application_id = $1 ORDER BY round`, appID,
	); err != nil {
		return nil, err
	}
	interviews := make([]domainapp.Interview, 0, len(rows))
	for _, row := range rows {
		iv := domainapp.Interview{
			ID:          row.ID,
			Round:       row.Round,
			Title:       row.Title,
			ScheduledAt: row.ScheduledAt,
			DurationMin: row.DurationMin,
			MeetingURL:  row.MeetingURL,
			Type:        row.Type,
		}
		_ = json.Unmarshal(row.InterviewerIDs, &iv.InterviewerIDs)
		if len(row.FeedbackJSON) > 0 {
			_ = json.Unmarshal(row.FeedbackJSON, &iv.Feedback)
		}
		interviews = append(interviews, iv)
	}
	return interviews, nil
}

func (r *applicationRepo) SaveOffer(ctx context.Context, appID uuid.UUID, offer domainapp.Offer) error {
	benefitsJSON, _ := json.Marshal(offer.Benefits)
	salaryJSON, _ := json.Marshal(offer.Salary)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO offers (id, application_id, salary, start_date, title, benefits, offer_letter_url, expires_at, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		offer.ID, appID, salaryJSON, offer.StartDate, offer.Title,
		benefitsJSON, offer.OfferLetterURL, offer.ExpiresAt, offer.SentAt,
	)
	return err
}

func (r *applicationRepo) findOffer(ctx context.Context, appID uuid.UUID) (*domainapp.Offer, error) {
	type offerRow struct {
		ID             uuid.UUID  `db:"id"`
		SalaryJSON     []byte     `db:"salary"`
		StartDate      time.Time  `db:"start_date"`
		Title          string     `db:"title"`
		BenefitsJSON   []byte     `db:"benefits"`
		OfferLetterURL string     `db:"offer_letter_url"`
		ExpiresAt      time.Time  `db:"expires_at"`
		SentAt         *time.Time `db:"sent_at"`
		RespondedAt    *time.Time `db:"responded_at"`
	}
	var row offerRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM offers WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1`, appID,
	); err != nil {
		return nil, err
	}
	o := &domainapp.Offer{
		ID:             row.ID,
		StartDate:      row.StartDate,
		Title:          row.Title,
		OfferLetterURL: row.OfferLetterURL,
		ExpiresAt:      row.ExpiresAt,
		SentAt:         row.SentAt,
		RespondedAt:    row.RespondedAt,
	}
	_ = json.Unmarshal(row.SalaryJSON, &o.Salary)
	_ = json.Unmarshal(row.BenefitsJSON, &o.Benefits)
	return o, nil
}

type appRow struct {
	ID                  uuid.UUID  `db:"id"`
	JobID               uuid.UUID  `db:"job_id"`
	CandidateID         uuid.UUID  `db:"candidate_id"`
	RecruiterID         uuid.UUID  `db:"recruiter_id"`
	Status              string     `db:"status"`
	CurrentStageID      *uuid.UUID `db:"current_stage_id"`
	RejectionReason     *string    `db:"rejection_reason"`
	WithdrawReason      string     `db:"withdraw_reason"`
	MatchScoreJSON      []byte     `db:"match_score"`
	Priority            int        `db:"priority"`
	ReferredByPartnerID *uuid.UUID `db:"referred_by_partner_id"`
	DaysInStage         int        `db:"days_in_stage"`
	LastMovedAt         time.Time  `db:"last_moved_at"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
}

func appToRow(a *domainapp.Application) map[string]interface{} {
	matchScoreJSON, _ := json.Marshal(a.MatchScore)
	return map[string]interface{}{
		"id":                     a.ID,
		"job_id":                 a.JobID,
		"candidate_id":           a.CandidateID,
		"recruiter_id":           a.RecruiterID,
		"status":                 a.Status,
		"current_stage_id":       a.CurrentStageID,
		"rejection_reason":       a.RejectionReason,
		"withdraw_reason":        a.WithdrawReason,
		"match_score":            matchScoreJSON,
		"priority":               a.Priority,
		"referred_by_partner_id": a.ReferredByPartnerID,
		"days_in_stage":          a.DaysInStage,
		"last_moved_at":          a.LastMovedAt,
		"created_at":             a.CreatedAt,
		"updated_at":             a.UpdatedAt,
	}
}

func appFromRow(row *appRow) (*domainapp.Application, error) {
	a := &domainapp.Application{}
	a.ID = row.ID
	a.JobID = row.JobID
	a.CandidateID = row.CandidateID
	a.RecruiterID = row.RecruiterID
	a.Status = domainapp.Status(row.Status)
	a.CurrentStageID = uuid.Nil
	if row.CurrentStageID != nil {
		a.CurrentStageID = *row.CurrentStageID
	}
	if row.RejectionReason != nil {
		r := domainapp.RejectionReason(*row.RejectionReason)
		a.RejectionReason = &r
	}
	a.WithdrawReason = row.WithdrawReason
	a.Priority = row.Priority
	a.ReferredByPartnerID = row.ReferredByPartnerID
	a.DaysInStage = row.DaysInStage
	a.LastMovedAt = row.LastMovedAt
	a.CreatedAt = row.CreatedAt
	a.UpdatedAt = row.UpdatedAt

	_ = json.Unmarshal(row.MatchScoreJSON, &a.MatchScore)
	return a, nil
}
