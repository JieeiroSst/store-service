package applicationusecase

import (
	"context"
	"errors"

	domainapp "github.com/JIeeiroSst/recruitment-platform-service/internal/domain/application"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/job"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type service struct {
	appRepo       domainapp.Repository
	jobRepo       job.Repository
	candidateRepo candidate.Repository
	aiSvc         port.AIService
	workflowSvc   port.WorkflowService
	eventBus      port.EventBus
	logger        *zap.Logger
}

func NewService(
	appRepo domainapp.Repository,
	jobRepo job.Repository,
	candidateRepo candidate.Repository,
	aiSvc port.AIService,
	workflowSvc port.WorkflowService,
	eventBus port.EventBus,
	logger *zap.Logger,
) port.ApplicationService {
	return &service{
		appRepo:       appRepo,
		jobRepo:       jobRepo,
		candidateRepo: candidateRepo,
		aiSvc:         aiSvc,
		workflowSvc:   workflowSvc,
		eventBus:      eventBus,
		logger:        logger,
	}
}

func (s *service) Apply(ctx context.Context, cmd port.ApplyCommand) (*domainapp.Application, error) {
	existing, err := s.appRepo.FindByJobAndCandidate(ctx, cmd.JobID, cmd.CandidateID)
	if err == nil && existing != nil {
		return nil, errors.New("application: already applied to this job")
	}

	j, err := s.jobRepo.FindByID(ctx, cmd.JobID)
	if err != nil {
		return nil, errors.New("application: job not found")
	}
	if j.Status != job.StatusOpen {
		return nil, errors.New("application: job is not open for applications")
	}

	app, err := domainapp.New(cmd.JobID, cmd.CandidateID, cmd.RecruiterID)
	if err != nil {
		return nil, err
	}

	if cmd.ReferredByPartnerID != nil {
		app.ReferredByPartnerID = cmd.ReferredByPartnerID
	}

	if err := s.appRepo.Save(ctx, app); err != nil {
		return nil, err
	}

	if err := s.workflowSvc.StartRecruitmentWorkflow(ctx, port.StartWorkflowInput{
		ApplicationID: app.ID,
		JobID:         app.JobID,
		CandidateID:   app.CandidateID,
	}); err != nil {
		s.logger.Warn("failed to start recruitment workflow", zap.Error(err), zap.String("app_id", app.ID.String()))
	}

	go func() {
		score, err := s.aiSvc.ScoreMatch(context.Background(), port.ScoreMatchInput{
			JobID:       app.JobID,
			CandidateID: app.CandidateID,
		})
		if err != nil {
			s.logger.Error("ai score match failed", zap.Error(err))
			return
		}
		app.MatchScore = &score
		if err := s.appRepo.Update(context.Background(), app); err != nil {
			s.logger.Error("failed to persist ai score", zap.Error(err))
		}
	}()

	s.publishEvents(ctx, app.DomainEvents())
	app.ClearEvents()
	return app, nil
}

func (s *service) MoveStage(ctx context.Context, cmd port.MoveStageCommand) (*domainapp.Application, error) {
	app, err := s.appRepo.FindByID(ctx, cmd.ApplicationID)
	if err != nil {
		return nil, err
	}
	if err := app.MoveToStage(cmd.StageID, domainapp.Status(cmd.Status)); err != nil {
		return nil, err
	}
	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, err
	}

	_ = s.workflowSvc.SignalStageChange(ctx, app.ID, string(app.Status))

	s.publishEvents(ctx, app.DomainEvents())
	app.ClearEvents()
	return app, nil
}

func (s *service) ScheduleInterview(ctx context.Context, cmd port.ScheduleInterviewCommand) (*domainapp.Application, error) {
	app, err := s.appRepo.FindByID(ctx, cmd.ApplicationID)
	if err != nil {
		return nil, err
	}

	interview := domainapp.Interview{
		ID:             uuid.New(),
		Round:          cmd.Round,
		Title:          cmd.Title,
		InterviewerIDs: cmd.InterviewerIDs,
		ScheduledAt:    cmd.ScheduledAt,
		DurationMin:    cmd.DurationMin,
		MeetingURL:     cmd.MeetingURL,
		Type:           cmd.Type,
	}
	app.AddInterview(interview)

	if err := s.appRepo.SaveInterview(ctx, app.ID, interview); err != nil {
		return nil, err
	}

	_ = s.workflowSvc.SignalInterviewScheduled(ctx, app.ID, interview.ID)

	s.publishEvents(ctx, app.DomainEvents())
	app.ClearEvents()
	return app, nil
}

func (s *service) SubmitFeedback(ctx context.Context, cmd port.SubmitFeedbackCommand) error {
	app, err := s.appRepo.FindByID(ctx, cmd.ApplicationID)
	if err != nil {
		return err
	}
	feedback := domainapp.InterviewFeedback{
		SubmittedBy: cmd.SubmittedBy,
		Decision:    cmd.Decision,
		Score:       cmd.Score,
		Strengths:   cmd.Strengths,
		Weaknesses:  cmd.Weaknesses,
		Notes:       cmd.Notes,
	}
	if err := app.SubmitFeedback(cmd.InterviewID, feedback); err != nil {
		return err
	}
	if err := s.appRepo.UpdateInterview(ctx, app.ID, domainapp.Interview{ID: cmd.InterviewID, Feedback: &feedback}); err != nil {
		return err
	}
	s.publishEvents(ctx, app.DomainEvents())
	app.ClearEvents()
	return nil
}

func (s *service) ExtendOffer(ctx context.Context, cmd port.ExtendOfferCommand) (*domainapp.Application, error) {
	app, err := s.appRepo.FindByID(ctx, cmd.ApplicationID)
	if err != nil {
		return nil, err
	}
	offer := domainapp.Offer{
		ID:        uuid.New(),
		Salary:    cmd.Salary,
		StartDate: cmd.StartDate,
		Title:     cmd.Title,
		Benefits:  cmd.Benefits,
		ExpiresAt: cmd.ExpiresAt,
	}
	if err := app.ExtendOffer(offer); err != nil {
		return nil, err
	}
	if err := s.appRepo.SaveOffer(ctx, app.ID, offer); err != nil {
		return nil, err
	}
	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, err
	}

	_ = s.workflowSvc.SignalOfferExtended(ctx, app.ID, offer.ID)

	s.publishEvents(ctx, app.DomainEvents())
	app.ClearEvents()
	return app, nil
}

func (s *service) Reject(ctx context.Context, cmd port.RejectCommand) error {
	app, err := s.appRepo.FindByID(ctx, cmd.ApplicationID)
	if err != nil {
		return err
	}
	if err := app.Reject(domainapp.RejectionReason(cmd.Reason), cmd.Note); err != nil {
		return err
	}
	if err := s.appRepo.Update(ctx, app); err != nil {
		return err
	}
	_ = s.workflowSvc.TerminateWorkflow(ctx, app.ID, "rejected: "+cmd.Note)
	s.publishEvents(ctx, app.DomainEvents())
	app.ClearEvents()
	return nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domainapp.Application, error) {
	return s.appRepo.FindByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter domainapp.Filter) (shared.PaginatedResult[*domainapp.Application], error) {
	return s.appRepo.FindAll(ctx, filter)
}

func (s *service) publishEvents(ctx context.Context, events []shared.DomainEvent) {
	for _, e := range events {
		if err := s.eventBus.Publish(ctx, e); err != nil {
			s.logger.Error("event publish failed", zap.String("type", e.EventType), zap.Error(err))
		}
	}
}
