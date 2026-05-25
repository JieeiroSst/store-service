package activity

import (
	"context"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotifyCandidateInput struct {
	CandidateID   uuid.UUID      `json:"candidate_id"`
	ApplicationID uuid.UUID      `json:"application_id"`
	TemplateID    string         `json:"template_id"`
	Data          map[string]any `json:"data,omitempty"`
}

type AIScoreInput struct {
	JobID       uuid.UUID `json:"job_id"`
	CandidateID uuid.UUID `json:"candidate_id"`
}

type AIScoreResult struct {
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

type SendInterviewInviteInput struct {
	CandidateID   uuid.UUID `json:"candidate_id"`
	ApplicationID uuid.UUID `json:"application_id"`
	InterviewID   uuid.UUID `json:"interview_id"`
	ScheduledAt   time.Time `json:"scheduled_at"`
}

type TriggerReferralPayoutInput struct {
	ApplicationID uuid.UUID `json:"application_id"`
	CandidateID   uuid.UUID `json:"candidate_id"`
}

type ComputeCommissionInput struct {
	PartnerID     uuid.UUID `json:"partner_id"`
	ApplicationID uuid.UUID `json:"application_id"`
}

type ComputeCommissionResult struct {
	CommissionAmount int64  `json:"commission_amount"`
	Currency         string `json:"currency"`
}

type Activities struct {
	notificationSvc port.NotificationService
	aiSvc           port.AIService
	referralSvc     port.ReferralService
	candidateSvc    port.CandidateService
	logger          *zap.Logger
}

func NewActivities(
	notificationSvc port.NotificationService,
	aiSvc port.AIService,
	referralSvc port.ReferralService,
	candidateSvc port.CandidateService,
	logger *zap.Logger,
) *Activities {
	return &Activities{
		notificationSvc: notificationSvc,
		aiSvc:           aiSvc,
		referralSvc:     referralSvc,
		candidateSvc:    candidateSvc,
		logger:          logger,
	}
}

func (a *Activities) NotifyCandidateActivity(ctx context.Context, input NotifyCandidateInput) error {
	a.logger.Info("sending notification",
		zap.String("candidate_id", input.CandidateID.String()),
		zap.String("template", input.TemplateID),
	)
	data := input.Data
	if data == nil {
		data = make(map[string]any)
	}
	data["application_id"] = input.ApplicationID
	return a.notificationSvc.Send(ctx, port.NotificationPayload{
		RecipientID: input.CandidateID,
		Channel:     "email",
		TemplateID:  input.TemplateID,
		Data:        data,
	})
}

func (a *Activities) ComputeAIMatchScoreActivity(ctx context.Context, input AIScoreInput) (AIScoreResult, error) {
	score, err := a.aiSvc.ScoreMatch(ctx, port.ScoreMatchInput{
		JobID:       input.JobID,
		CandidateID: input.CandidateID,
	})
	if err != nil {
		return AIScoreResult{}, err
	}
	return AIScoreResult{Score: score.Score, Confidence: score.Confidence}, nil
}

func (a *Activities) SendInterviewInviteActivity(ctx context.Context, input SendInterviewInviteInput) error {
	return a.notificationSvc.Send(ctx, port.NotificationPayload{
		RecipientID: input.CandidateID,
		Channel:     "email",
		TemplateID:  "interview_invite",
		Data: map[string]any{
			"interview_id":   input.InterviewID,
			"application_id": input.ApplicationID,
			"scheduled_at":   input.ScheduledAt.Format(time.RFC3339),
		},
	})
}

func (a *Activities) TriggerReferralPayoutActivity(ctx context.Context, input TriggerReferralPayoutInput) error {
	a.logger.Info("triggering referral payout", zap.String("application_id", input.ApplicationID.String()))
	return a.referralSvc.TrackHire(ctx, input.ApplicationID)
}

func (a *Activities) ParseAndEnrichCandidateActivity(ctx context.Context, candidateID uuid.UUID) error {
	a.logger.Info("enriching candidate with AI", zap.String("candidate_id", candidateID.String()))
	return a.candidateSvc.EnrichWithAI(ctx, candidateID)
}

func (a *Activities) ComputeReferralCommissionActivity(
	ctx context.Context,
	input ComputeCommissionInput,
) (ComputeCommissionResult, error) {
	stats, err := a.referralSvc.GetPartnerStats(ctx, input.PartnerID)
	if err != nil {
		return ComputeCommissionResult{}, err
	}
	var amount int64
	switch stats.Partner.Tier {
	case "platinum":
		amount = 20_000_000
	case "gold":
		amount = 15_000_000
	case "silver":
		amount = 10_000_000
	default:
		amount = 5_000_000
	}
	return ComputeCommissionResult{CommissionAmount: amount, Currency: "VND"}, nil
}
