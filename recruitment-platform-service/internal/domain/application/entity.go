package application

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
)

type Status string

const (
	StatusApplied       Status = "applied"
	StatusCVReview      Status = "cv_review"
	StatusPhoneScreen   Status = "phone_screen"
	StatusTechnical     Status = "technical"
	StatusFinalRound    Status = "final_round"
	StatusOffer         Status = "offer"
	StatusOfferAccepted Status = "offer_accepted"
	StatusOfferDeclined Status = "offer_declined"
	StatusHired         Status = "hired"
	StatusRejected      Status = "rejected"
	StatusWithdrawn     Status = "withdrawn"
)

type RejectionReason string

const (
	RejectionSkillMismatch  RejectionReason = "skill_mismatch"
	RejectionSalaryMismatch RejectionReason = "salary_mismatch"
	RejectionCultureFit     RejectionReason = "culture_fit"
	RejectionOverqualified  RejectionReason = "overqualified"
	RejectionOtherOffer     RejectionReason = "other_offer"
	RejectionNoShow         RejectionReason = "no_show"
)

type Interview struct {
	ID             uuid.UUID          `json:"id"`
	Round          int                `json:"round"`
	Title          string             `json:"title"`
	InterviewerIDs []uuid.UUID        `json:"interviewer_ids"`
	ScheduledAt    time.Time          `json:"scheduled_at"`
	DurationMin    int                `json:"duration_min"`
	MeetingURL     string             `json:"meeting_url"`
	Type           string             `json:"type"` // online | onsite
	Feedback       *InterviewFeedback `json:"feedback,omitempty"`
}

type InterviewFeedback struct {
	SubmittedBy uuid.UUID `json:"submitted_by"`
	Decision    string    `json:"decision"` // pass | fail | hold
	Score       int       `json:"score"`    // 1–5
	Strengths   string    `json:"strengths"`
	Weaknesses  string    `json:"weaknesses"`
	Notes       string    `json:"notes"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type Offer struct {
	ID             uuid.UUID    `json:"id"`
	Salary         shared.Money `json:"salary"`
	StartDate      time.Time    `json:"start_date"`
	Title          string       `json:"title"`
	Benefits       []string     `json:"benefits"`
	OfferLetterURL string       `json:"offer_letter_url"`
	ExpiresAt      time.Time    `json:"expires_at"`
	SentAt         *time.Time   `json:"sent_at"`
	RespondedAt    *time.Time   `json:"responded_at"`
}

type Application struct {
	shared.BaseEntity

	JobID       uuid.UUID `db:"job_id"       json:"job_id"`
	CandidateID uuid.UUID `db:"candidate_id" json:"candidate_id"`
	RecruiterID uuid.UUID `db:"recruiter_id" json:"recruiter_id"`

	Status          Status           `db:"status"           json:"status"`
	CurrentStageID  uuid.UUID        `db:"current_stage_id" json:"current_stage_id"`
	RejectionReason *RejectionReason `db:"rejection_reason" json:"rejection_reason,omitempty"`
	WithdrawReason  string           `db:"withdraw_reason"  json:"withdraw_reason,omitempty"`

	Interviews []Interview `db:"-" json:"interviews,omitempty"`
	Offer      *Offer      `db:"-" json:"offer,omitempty"`

	MatchScore *shared.AIScore `db:"match_score" json:"match_score,omitempty"`
	Priority   int             `db:"priority"    json:"priority"`

	ReferredByPartnerID *uuid.UUID `db:"referred_by_partner_id" json:"referred_by_partner_id,omitempty"`

	DaysInStage int       `db:"days_in_stage" json:"days_in_stage"`
	LastMovedAt time.Time `db:"last_moved_at" json:"last_moved_at"`

	events []shared.DomainEvent
}

func New(jobID, candidateID, recruiterID uuid.UUID) (*Application, error) {
	if jobID == uuid.Nil || candidateID == uuid.Nil {
		return nil, errors.New("application: job and candidate IDs required")
	}
	now := time.Now()
	app := &Application{
		BaseEntity:  shared.NewBaseEntity(),
		JobID:       jobID,
		CandidateID: candidateID,
		RecruiterID: recruiterID,
		Status:      StatusApplied,
		LastMovedAt: now,
	}
	app.record("ApplicationCreated", map[string]any{
		"application_id": app.ID,
		"job_id":         jobID,
		"candidate_id":   candidateID,
	})
	return app, nil
}

func (a *Application) MoveToStage(stageID uuid.UUID, status Status) error {
	if a.Status == StatusHired || a.Status == StatusRejected || a.Status == StatusWithdrawn {
		return errors.New("application: terminal status, cannot move stage")
	}
	a.CurrentStageID = stageID
	a.Status = status
	a.DaysInStage = 0
	a.LastMovedAt = time.Now()
	a.record("ApplicationStageMoved", map[string]any{
		"application_id": a.ID,
		"new_stage":      stageID,
		"new_status":     status,
	})
	return nil
}

func (a *Application) Reject(reason RejectionReason, note string) error {
	if a.Status == StatusHired {
		return errors.New("application: cannot reject a hired application")
	}
	a.Status = StatusRejected
	a.RejectionReason = &reason
	a.record("ApplicationRejected", map[string]any{
		"application_id": a.ID,
		"reason":         reason,
		"note":           note,
	})
	return nil
}

func (a *Application) Withdraw(reason string) {
	a.Status = StatusWithdrawn
	a.WithdrawReason = reason
	a.record("ApplicationWithdrawn", map[string]any{"application_id": a.ID})
}

func (a *Application) ExtendOffer(offer Offer) error {
	if a.Status != StatusFinalRound {
		return errors.New("application: must be in final round to extend offer")
	}
	a.Offer = &offer
	a.Status = StatusOffer
	now := time.Now()
	a.Offer.SentAt = &now
	a.record("OfferExtended", map[string]any{"application_id": a.ID, "offer_id": offer.ID})
	return nil
}

func (a *Application) AcceptOffer() error {
	if a.Status != StatusOffer {
		return errors.New("application: no pending offer to accept")
	}
	a.Status = StatusOfferAccepted
	now := time.Now()
	if a.Offer != nil {
		a.Offer.RespondedAt = &now
	}
	a.record("OfferAccepted", map[string]any{"application_id": a.ID})
	return nil
}

func (a *Application) MarkHired() error {
	if a.Status != StatusOfferAccepted {
		return errors.New("application: offer must be accepted before marking hired")
	}
	a.Status = StatusHired
	a.record("CandidateHired", map[string]any{"application_id": a.ID})
	return nil
}

func (a *Application) AddInterview(interview Interview) {
	a.Interviews = append(a.Interviews, interview)
	a.record("InterviewScheduled", map[string]any{
		"application_id": a.ID,
		"interview_id":   interview.ID,
		"scheduled_at":   interview.ScheduledAt,
	})
}

func (a *Application) SubmitFeedback(interviewID uuid.UUID, feedback InterviewFeedback) error {
	for i, iv := range a.Interviews {
		if iv.ID == interviewID {
			a.Interviews[i].Feedback = &feedback
			a.record("InterviewFeedbackSubmitted", map[string]any{
				"application_id": a.ID,
				"interview_id":   interviewID,
				"decision":       feedback.Decision,
			})
			return nil
		}
	}
	return errors.New("application: interview not found")
}

func (a *Application) DomainEvents() []shared.DomainEvent { return a.events }
func (a *Application) ClearEvents()                       { a.events = nil }
func (a *Application) record(t string, p interface{}) {
	a.events = append(a.events, shared.NewDomainEvent(t, p))
}

type Filter struct {
	JobID       *uuid.UUID `form:"job_id"`
	CandidateID *uuid.UUID `form:"candidate_id"`
	RecruiterID *uuid.UUID `form:"recruiter_id"`
	Status      *Status    `form:"status"`
	shared.PaginationParams
}

type Repository interface {
	Save(ctx context.Context, a *Application) error
	Update(ctx context.Context, a *Application) error
	FindByID(ctx context.Context, id uuid.UUID) (*Application, error)
	FindAll(ctx context.Context, filter Filter) (shared.PaginatedResult[*Application], error)
	FindByJobAndCandidate(ctx context.Context, jobID, candidateID uuid.UUID) (*Application, error)
	SaveInterview(ctx context.Context, appID uuid.UUID, iv Interview) error
	UpdateInterview(ctx context.Context, appID uuid.UUID, iv Interview) error
	SaveOffer(ctx context.Context, appID uuid.UUID, offer Offer) error
}
