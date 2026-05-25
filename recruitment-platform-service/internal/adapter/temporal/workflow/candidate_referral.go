package workflow

import (
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/temporal/activity"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type CandidateLifecycleInput struct {
	CandidateID uuid.UUID `json:"candidate_id"`
}

func CandidateLifecycleWorkflow(ctx workflow.Context, input CandidateLifecycleInput) error {
	logger := workflow.GetLogger(ctx)

	actOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	})

	_ = workflow.ExecuteActivity(actOpts,
		(*activity.Activities).ParseAndEnrichCandidateActivity,
		input.CandidateID,
	).Get(ctx, nil)

	idlePeriod := 30 * 24 * time.Hour
	for {
		statusSignalCh := workflow.GetSignalChannel(ctx, "candidate_status_change")
		selector := workflow.NewSelector(ctx)

		var terminated bool
		var newStatus string

		timer := workflow.NewTimer(ctx, idlePeriod)
		selector.AddFuture(timer, func(f workflow.Future) {
			logger.Info("candidate idle, sending nurture email",
				"candidate_id", input.CandidateID,
			)
			_ = workflow.ExecuteActivity(actOpts,
				(*activity.Activities).NotifyCandidateActivity,
				activity.NotifyCandidateInput{
					CandidateID: input.CandidateID,
					TemplateID:  "candidate_nurture",
				},
			).Get(ctx, nil)
		})

		selector.AddReceive(statusSignalCh, func(ch workflow.ReceiveChannel, more bool) {
			ch.Receive(ctx, &newStatus)
			if newStatus == "hired" || newStatus == "blacklist" {
				terminated = true
			}
		})

		selector.Select(ctx)

		if terminated {
			logger.Info("candidate lifecycle workflow ending", "status", newStatus)
			return nil
		}
	}
}

const SignalHireConfirmed = "hire_confirmed"

type ReferralNetworkInput struct {
	ApplicationID uuid.UUID `json:"application_id"`
	CandidateID   uuid.UUID `json:"candidate_id"`
	PartnerID     uuid.UUID `json:"partner_id"`
}

type HireConfirmedSignal struct {
	ProbationPassed bool `json:"probation_passed"`
}

func ReferralNetworkWorkflow(ctx workflow.Context, input ReferralNetworkInput) error {
	logger := workflow.GetLogger(ctx)

	actOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    5,
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2.0,
		},
	})

	hireCh := workflow.GetSignalChannel(ctx, SignalHireConfirmed)
	timeoutTimer := workflow.NewTimer(ctx, 60*24*time.Hour)
	selector := workflow.NewSelector(ctx)

	var hireSignal HireConfirmedSignal
	var timedOut bool

	selector.AddFuture(timeoutTimer, func(f workflow.Future) {
		timedOut = true
	})
	selector.AddReceive(hireCh, func(ch workflow.ReceiveChannel, more bool) {
		ch.Receive(ctx, &hireSignal)
	})
	selector.Select(ctx)

	if timedOut {
		logger.Info("referral workflow expired", "application_id", input.ApplicationID)
		return nil
	}

	if !hireSignal.ProbationPassed {
		logger.Info("waiting for probation", "partner_id", input.PartnerID)
		probationCh := workflow.GetSignalChannel(ctx, "probation_confirmed")
		probSelector := workflow.NewSelector(ctx)
		probTimer := workflow.NewTimer(ctx, 90*24*time.Hour)

		var probationPassed bool
		probSelector.AddFuture(probTimer, func(f workflow.Future) {})
		probSelector.AddReceive(probationCh, func(ch workflow.ReceiveChannel, more bool) {
			ch.Receive(ctx, &probationPassed)
		})
		probSelector.Select(ctx)

		if !probationPassed {
			logger.Info("probation not confirmed, skipping payout")
			return nil
		}
	}

	var commission activity.ComputeCommissionResult
	_ = workflow.ExecuteActivity(actOpts,
		(*activity.Activities).ComputeReferralCommissionActivity,
		activity.ComputeCommissionInput{
			PartnerID:     input.PartnerID,
			ApplicationID: input.ApplicationID,
		},
	).Get(ctx, &commission)

	_ = workflow.ExecuteActivity(actOpts,
		(*activity.Activities).TriggerReferralPayoutActivity,
		activity.TriggerReferralPayoutInput{
			ApplicationID: input.ApplicationID,
			CandidateID:   input.CandidateID,
		},
	).Get(ctx, nil)

	_ = workflow.ExecuteActivity(actOpts,
		(*activity.Activities).NotifyCandidateActivity,
		activity.NotifyCandidateInput{
			CandidateID: input.PartnerID, 
			TemplateID:  "referral_commission_ready",
			Data: map[string]any{
				"commission_amount": commission.CommissionAmount,
				"currency":          commission.Currency,
			},
		},
	).Get(ctx, nil)

	logger.Info("referral workflow completed",
		"partner_id", input.PartnerID,
		"commission", commission.CommissionAmount,
	)
	return nil
}
