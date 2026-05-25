package temporaladapter

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/temporal/activity"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/temporal/workflow"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

type Worker struct {
	temporalClient client.Client
	worker         worker.Worker
	activities     *activity.Activities
	logger         *zap.Logger
}

func NewWorker(
	tc client.Client,
	acts *activity.Activities,
	logger *zap.Logger,
) *Worker {
	return &Worker{
		temporalClient: tc,
		activities:     acts,
		logger:         logger,
	}
}

func (w *Worker) Start() error {
	w.worker = worker.New(w.temporalClient, workflow.TaskQueueRecruitment, worker.Options{
		MaxConcurrentActivityExecutionSize:     100,
		MaxConcurrentWorkflowTaskExecutionSize: 50,
	})

	// Register workflows
	w.worker.RegisterWorkflow(workflow.RecruitmentLifecycleWorkflow)
	w.worker.RegisterWorkflow(workflow.CandidateLifecycleWorkflow)
	w.worker.RegisterWorkflow(workflow.ReferralNetworkWorkflow)

	// Register activities (struct-based, sharing injected deps)
	w.worker.RegisterActivity(w.activities)

	w.logger.Info("temporal worker starting", zap.String("task_queue", workflow.TaskQueueRecruitment))
	return w.worker.Start()
}

func (w *Worker) Stop() {
	if w.worker != nil {
		w.worker.Stop()
	}
}

type workflowServiceAdapter struct {
	client client.Client
	logger *zap.Logger
}

func NewWorkflowService(c client.Client, logger *zap.Logger) port.WorkflowService {
	return &workflowServiceAdapter{client: c, logger: logger}
}

func (s *workflowServiceAdapter) StartRecruitmentWorkflow(ctx context.Context, input port.StartWorkflowInput) error {
	workflowID := fmt.Sprintf("recruitment-%s", input.ApplicationID)
	_, err := s.client.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: workflow.TaskQueueRecruitment,
		},
		workflow.RecruitmentLifecycleWorkflow,
		workflow.RecruitmentWorkflowInput{
			ApplicationID: input.ApplicationID,
			JobID:         input.JobID,
			CandidateID:   input.CandidateID,
		},
	)
	if err != nil {
		s.logger.Error("failed to start recruitment workflow", zap.Error(err))
		return err
	}
	return nil
}

func (s *workflowServiceAdapter) SignalStageChange(ctx context.Context, applicationID uuid.UUID, newStatus string) error {
	workflowID := fmt.Sprintf("recruitment-%s", applicationID)
	return s.client.SignalWorkflow(ctx, workflowID, "", workflow.SignalStageChange,
		workflow.StageChangeSignal{NewStatus: newStatus},
	)
}

func (s *workflowServiceAdapter) SignalInterviewScheduled(ctx context.Context, applicationID, interviewID uuid.UUID) error {
	workflowID := fmt.Sprintf("recruitment-%s", applicationID)
	return s.client.SignalWorkflow(ctx, workflowID, "", workflow.SignalInterviewScheduled,
		workflow.InterviewScheduledSignal{InterviewID: interviewID},
	)
}

func (s *workflowServiceAdapter) SignalOfferExtended(ctx context.Context, applicationID, offerID uuid.UUID) error {
	workflowID := fmt.Sprintf("recruitment-%s", applicationID)
	return s.client.SignalWorkflow(ctx, workflowID, "", workflow.SignalOfferExtended,
		workflow.OfferExtendedSignal{OfferID: offerID},
	)
}

func (s *workflowServiceAdapter) TerminateWorkflow(ctx context.Context, applicationID uuid.UUID, reason string) error {
	workflowID := fmt.Sprintf("recruitment-%s", applicationID)
	return s.client.TerminateWorkflow(ctx, workflowID, "", reason)
}
