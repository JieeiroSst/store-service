package workflow

import (
	"errors"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/temporal/activity"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	SignalStageChange        = "stage_change"
	SignalInterviewScheduled = "interview_scheduled"
	SignalOfferExtended      = "offer_extended"
	SignalOfferResponse      = "offer_response"
	QueryCurrentStatus       = "current_status"

	TaskQueueRecruitment = "recruitment-task-queue"
)


type RecruitmentWorkflowInput struct {
	ApplicationID uuid.UUID `json:"application_id"`
	JobID         uuid.UUID `json:"job_id"`
	CandidateID   uuid.UUID `json:"candidate_id"`
}

type StageChangeSignal struct {
	NewStatus string `json:"new_status"`
}

type InterviewScheduledSignal struct {
	InterviewID uuid.UUID `json:"interview_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

type OfferExtendedSignal struct {
	OfferID   uuid.UUID `json:"offer_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type OfferResponseSignal struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason"`
}


func RecruitmentLifecycleWorkflow(ctx workflow.Context, input RecruitmentWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("RecruitmentLifecycleWorkflow started",
		"application_id", input.ApplicationID,
		"job_id", input.JobID,
		"candidate_id", input.CandidateID,
	)

	actOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: 2 * time.Second,
		},
	})

	if err := workflow.ExecuteActivity(actOpts,
		(*activity.Activities).NotifyCandidateActivity,
		activity.NotifyCandidateInput{
			CandidateID:   input.CandidateID,
			ApplicationID: input.ApplicationID,
			TemplateID:    "application_received",
		},
	).Get(ctx, nil); err != nil {
		logger.Error("notify candidate failed", "err", err)
	}

	var aiScoreResult activity.AIScoreResult
	if err := workflow.ExecuteActivity(actOpts,
		(*activity.Activities).ComputeAIMatchScoreActivity,
		activity.AIScoreInput{
			JobID:       input.JobID,
			CandidateID: input.CandidateID,
		},
	).Get(ctx, &aiScoreResult); err != nil {
		logger.Error("ai score failed", "err", err)
	}

	stageSignalCh := workflow.GetSignalChannel(ctx, SignalStageChange)
	interviewSignalCh := workflow.GetSignalChannel(ctx, SignalInterviewScheduled)
	offerSignalCh := workflow.GetSignalChannel(ctx, SignalOfferExtended)
	offerResponseCh := workflow.GetSignalChannel(ctx, SignalOfferResponse)

	currentStatus := "applied"
	_ = workflow.SetQueryHandler(ctx, QueryCurrentStatus, func() (string, error) {
		return currentStatus, nil
	})

	for {
		var terminateReason string

		selector := workflow.NewSelector(ctx)
		timeoutTimer := workflow.NewTimer(ctx, 90*24*time.Hour) // 90-day SLA

		selector.AddFuture(timeoutTimer, func(f workflow.Future) {
			terminateReason = "sla_timeout"
		})

		selector.AddReceive(stageSignalCh, func(ch workflow.ReceiveChannel, more bool) {
			var sig StageChangeSignal
			ch.Receive(ctx, &sig)
			logger.Info("stage changed", "new_status", sig.NewStatus)
			currentStatus = sig.NewStatus

			switch sig.NewStatus {
			case "rejected", "withdrawn":
				terminateReason = sig.NewStatus
			case "cv_review":
				_ = workflow.ExecuteActivity(actOpts,
					(*activity.Activities).NotifyCandidateActivity,
					activity.NotifyCandidateInput{
						CandidateID:   input.CandidateID,
						ApplicationID: input.ApplicationID,
						TemplateID:    "cv_under_review",
					},
				).Get(ctx, nil)
			}
		})

		selector.AddReceive(interviewSignalCh, func(ch workflow.ReceiveChannel, more bool) {
			var sig InterviewScheduledSignal
			ch.Receive(ctx, &sig)
			currentStatus = "interview_scheduled"
			logger.Info("interview scheduled", "interview_id", sig.InterviewID)

			_ = workflow.ExecuteActivity(actOpts,
				(*activity.Activities).SendInterviewInviteActivity,
				activity.SendInterviewInviteInput{
					CandidateID:   input.CandidateID,
					ApplicationID: input.ApplicationID,
					InterviewID:   sig.InterviewID,
					ScheduledAt:   sig.ScheduledAt,
				},
			).Get(ctx, nil)

			reminderDelay := sig.ScheduledAt.Add(-24 * time.Hour).Sub(workflow.Now(ctx))
			if reminderDelay > 0 {
				workflow.Go(ctx, func(gCtx workflow.Context) {
					_ = workflow.NewTimer(gCtx, reminderDelay).Get(gCtx, nil)
					_ = workflow.ExecuteActivity(
						workflow.WithActivityOptions(gCtx, workflow.ActivityOptions{
							StartToCloseTimeout: 30 * time.Second,
						}),
						(*activity.Activities).NotifyCandidateActivity,
						activity.NotifyCandidateInput{
							CandidateID:   input.CandidateID,
							ApplicationID: input.ApplicationID,
							TemplateID:    "interview_reminder_24h",
						},
					).Get(gCtx, nil)
				})
			}
		})

		selector.AddReceive(offerSignalCh, func(ch workflow.ReceiveChannel, more bool) {
			var sig OfferExtendedSignal
			ch.Receive(ctx, &sig)
			currentStatus = "offer_extended"
			_ = workflow.ExecuteActivity(actOpts,
				(*activity.Activities).NotifyCandidateActivity,
				activity.NotifyCandidateInput{
					CandidateID:   input.CandidateID,
					ApplicationID: input.ApplicationID,
					TemplateID:    "offer_extended",
				},
			).Get(ctx, nil)
		})

		selector.AddReceive(offerResponseCh, func(ch workflow.ReceiveChannel, more bool) {
			var sig OfferResponseSignal
			ch.Receive(ctx, &sig)
			if sig.Accepted {
				currentStatus = "offer_accepted"
			} else {
				currentStatus = "offer_declined"
				terminateReason = "offer_declined"
			}
		})

		selector.Select(ctx)

		if terminateReason != "" {
			if currentStatus == "hired" || currentStatus == "offer_accepted" {
				_ = workflow.ExecuteActivity(actOpts,
					(*activity.Activities).TriggerReferralPayoutActivity,
					activity.TriggerReferralPayoutInput{
						ApplicationID: input.ApplicationID,
						CandidateID:   input.CandidateID,
					},
				).Get(ctx, nil)
			}
			if terminateReason == "sla_timeout" {
				return errors.New("workflow: SLA timeout exceeded for application " + input.ApplicationID.String())
			}
			logger.Info("workflow terminated", "reason", terminateReason)
			return nil
		}
	}
}
