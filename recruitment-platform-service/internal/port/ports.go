package port

import (
	"context"
	"time"

	domainapp "github.com/JIeeiroSst/recruitment-platform-service/internal/domain/application"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/job"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/referral"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
)

type CreateCandidateCommand struct {
	FullName      string
	Email         string
	Phone         string
	Source        candidate.SourceChannel
	ReferralToken string
	ResumeURL     string
}

type CandidateService interface {
	Create(ctx context.Context, cmd CreateCandidateCommand) (*candidate.Candidate, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*candidate.Candidate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*candidate.Candidate, error)
	List(ctx context.Context, filter candidate.Filter) (shared.PaginatedResult[*candidate.Candidate], error)
	Transition(ctx context.Context, id uuid.UUID, status candidate.Status, reason string) error
	EnrichWithAI(ctx context.Context, id uuid.UUID) error // parse resume → update skills/embedding
}

type CreateJobCommand struct {
	Title           string
	Code            string
	DepartmentID    uuid.UUID
	HiringManagerID uuid.UUID
	RecruiterIDs    []uuid.UUID
	Description     string
	Requirements    []string
	Skills          []string
	JobType         job.JobType
	WorkMode        job.WorkMode
	Location        shared.Address
	SalaryMin       *shared.Money
	SalaryMax       *shared.Money
	Headcount       int
}

type JobService interface {
	Create(ctx context.Context, cmd CreateJobCommand) (*job.Job, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*job.Job, error)
	Publish(ctx context.Context, id uuid.UUID) error
	Pause(ctx context.Context, id uuid.UUID) error
	Close(ctx context.Context, id uuid.UUID, reason string) error
	GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error)
	List(ctx context.Context, filter job.Filter) (shared.PaginatedResult[*job.Job], error)
	GetRecommendedCandidates(ctx context.Context, jobID uuid.UUID, limit int) ([]*candidate.Candidate, error)
}

type ApplyCommand struct {
	JobID               uuid.UUID
	CandidateID         uuid.UUID
	RecruiterID         uuid.UUID
	ReferredByPartnerID *uuid.UUID
	CoverLetter         string
}

type MoveStageCommand struct {
	ApplicationID uuid.UUID
	StageID       uuid.UUID
	Status        string
}

type ScheduleInterviewCommand struct {
	ApplicationID  uuid.UUID
	Round          int
	Title          string
	InterviewerIDs []uuid.UUID
	ScheduledAt    time.Time
	DurationMin    int
	MeetingURL     string
	Type           string
}

type SubmitFeedbackCommand struct {
	ApplicationID uuid.UUID
	InterviewID   uuid.UUID
	SubmittedBy   uuid.UUID
	Decision      string
	Score         int
	Strengths     string
	Weaknesses    string
	Notes         string
}

type ExtendOfferCommand struct {
	ApplicationID uuid.UUID
	Salary        shared.Money
	StartDate     time.Time
	Title         string
	Benefits      []string
	ExpiresAt     time.Time
}

type RejectCommand struct {
	ApplicationID uuid.UUID
	Reason        string
	Note          string
}

type ApplicationService interface {
	Apply(ctx context.Context, cmd ApplyCommand) (*domainapp.Application, error)
	MoveStage(ctx context.Context, cmd MoveStageCommand) (*domainapp.Application, error)
	ScheduleInterview(ctx context.Context, cmd ScheduleInterviewCommand) (*domainapp.Application, error)
	SubmitFeedback(ctx context.Context, cmd SubmitFeedbackCommand) error
	ExtendOffer(ctx context.Context, cmd ExtendOfferCommand) (*domainapp.Application, error)
	Reject(ctx context.Context, cmd RejectCommand) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainapp.Application, error)
	List(ctx context.Context, filter domainapp.Filter) (shared.PaginatedResult[*domainapp.Application], error)
}

type RegisterPartnerCommand struct {
	UserID          uuid.UUID
	FullName        string
	Email           string
	Phone           string
	Company         string
	ReferredByToken string 
}

type ReferralService interface {
	RegisterPartner(ctx context.Context, cmd RegisterPartnerCommand) (*referral.Partner, error)
	GenerateReferralLink(ctx context.Context, partnerID uuid.UUID, jobID *uuid.UUID) (*referral.Referral, error)
	TrackReferralClick(ctx context.Context, token string) error
	TrackApplication(ctx context.Context, token string, candidateID, applicationID uuid.UUID) error
	TrackHire(ctx context.Context, applicationID uuid.UUID) error
	GetPartnerStats(ctx context.Context, partnerID uuid.UUID) (*PartnerStats, error)
	GetLeaderboard(ctx context.Context, limit int) ([]*PartnerStats, error)
	RequestPayout(ctx context.Context, partnerID uuid.UUID) (*referral.Payout, error)
}

type PartnerStats struct {
	Partner        *referral.Partner `json:"partner"`
	TotalReferrals int               `json:"total_referrals"`
	HiredCount     int               `json:"hired_count"`
	ConversionRate float64           `json:"conversion_rate"`
	PendingPayout  shared.Money      `json:"pending_payout"`
	TotalEarned    shared.Money      `json:"total_earned"`
	Rank           int               `json:"rank"`
}

type ScoreMatchInput struct {
	JobID       uuid.UUID
	CandidateID uuid.UUID
}

type ParseResumeOutput struct {
	Skills          []string
	YearsExperience int
	ExperienceLevel candidate.ExperienceLevel
	Embedding       []float32
	SuggestedTitle  string
}

type AIService interface {
	ScoreMatch(ctx context.Context, input ScoreMatchInput) (shared.AIScore, error)
	ParseResume(ctx context.Context, resumeURL string) (*ParseResumeOutput, error)
	RecommendCandidates(ctx context.Context, jobID uuid.UUID, limit int) ([]*candidate.Candidate, error)
	GenerateJobDescription(ctx context.Context, title, requirements string) (string, error)
}

type StartWorkflowInput struct {
	ApplicationID uuid.UUID
	JobID         uuid.UUID
	CandidateID   uuid.UUID
}

type WorkflowService interface {
	StartRecruitmentWorkflow(ctx context.Context, input StartWorkflowInput) error
	SignalStageChange(ctx context.Context, applicationID uuid.UUID, newStatus string) error
	SignalInterviewScheduled(ctx context.Context, applicationID, interviewID uuid.UUID) error
	SignalOfferExtended(ctx context.Context, applicationID, offerID uuid.UUID) error
	TerminateWorkflow(ctx context.Context, applicationID uuid.UUID, reason string) error
}

type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
	Subscribe(eventType string, handler func(context.Context, shared.DomainEvent) error)
}

type NotificationPayload struct {
	RecipientID uuid.UUID
	Channel     string // "email" | "sms" | "push"
	TemplateID  string
	Data        map[string]any
}

type NotificationService interface {
	Send(ctx context.Context, n NotificationPayload) error
	SendBulk(ctx context.Context, ns []NotificationPayload) error
}
