package jobusecase

import (
	"context"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/job"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type service struct {
	jobRepo  job.Repository
	aiSvc    port.AIService
	eventBus port.EventBus
	logger   *zap.Logger
}

func NewService(
	jobRepo job.Repository,
	aiSvc port.AIService,
	eventBus port.EventBus,
	logger *zap.Logger,
) port.JobService {
	return &service{
		jobRepo:  jobRepo,
		aiSvc:    aiSvc,
		eventBus: eventBus,
		logger:   logger,
	}
}

func (s *service) Create(ctx context.Context, cmd port.CreateJobCommand) (*job.Job, error) {
	j, err := job.New(cmd.Title, cmd.Code, cmd.DepartmentID, cmd.HiringManagerID)
	if err != nil {
		return nil, err
	}
	j.RecruiterIDs = cmd.RecruiterIDs
	j.Description = cmd.Description
	j.Requirements = cmd.Requirements
	j.Skills = cmd.Skills
	j.JobType = cmd.JobType
	j.WorkMode = cmd.WorkMode
	j.Location = cmd.Location
	j.SalaryMin = cmd.SalaryMin
	j.SalaryMax = cmd.SalaryMax
	j.Headcount = cmd.Headcount

	if err := s.jobRepo.Save(ctx, j); err != nil {
		return nil, err
	}

	go func() {
		desc := cmd.Title + " " + cmd.Description
		result, err := s.aiSvc.ParseResume(context.Background(), desc) // reused for text→embedding
		if err != nil {
			s.logger.Warn("job embedding failed", zap.String("job_id", j.ID.String()), zap.Error(err))
			return
		}
		j.Embedding = result.Embedding
		if err := s.jobRepo.Update(context.Background(), j); err != nil {
			s.logger.Error("job embedding save failed", zap.Error(err))
		}
	}()

	return j, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*job.Job, error) {
	j, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := updates["title"].(string); ok {
		j.Title = v
	}
	if v, ok := updates["description"].(string); ok {
		j.Description = v
	}
	if v, ok := updates["requirements"].([]string); ok {
		j.Requirements = v
	}
	if v, ok := updates["skills"].([]string); ok {
		j.Skills = v
	}
	if v, ok := updates["headcount"].(int); ok {
		j.Headcount = v
	}
	if v, ok := updates["priority"].(int); ok {
		j.Priority = v
	}
	if v, ok := updates["work_mode"].(job.WorkMode); ok {
		j.WorkMode = v
	}

	if err := s.jobRepo.Update(ctx, j); err != nil {
		return nil, err
	}
	return j, nil
}

func (s *service) Publish(ctx context.Context, id uuid.UUID) error {
	j, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := j.Publish(); err != nil {
		return err
	}
	if err := s.jobRepo.Update(ctx, j); err != nil {
		return err
	}
	s.publishEvents(ctx, j.DomainEvents())
	j.ClearEvents()
	return nil
}

func (s *service) Pause(ctx context.Context, id uuid.UUID) error {
	j, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := j.Pause(); err != nil {
		return err
	}
	return s.jobRepo.Update(ctx, j)
}

func (s *service) Close(ctx context.Context, id uuid.UUID, reason string) error {
	j, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	j.Close(reason)
	if err := s.jobRepo.Update(ctx, j); err != nil {
		return err
	}
	s.publishEvents(ctx, j.DomainEvents())
	j.ClearEvents()
	return nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	return s.jobRepo.FindByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter job.Filter) (shared.PaginatedResult[*job.Job], error) {
	return s.jobRepo.FindAll(ctx, filter)
}

func (s *service) GetRecommendedCandidates(ctx context.Context, jobID uuid.UUID, limit int) ([]*candidate.Candidate, error) {
	return s.aiSvc.RecommendCandidates(ctx, jobID, limit)
}

func (s *service) publishEvents(ctx context.Context, events []shared.DomainEvent) {
	for _, e := range events {
		if err := s.eventBus.Publish(ctx, e); err != nil {
			s.logger.Error("event publish failed", zap.String("type", e.EventType), zap.Error(err))
		}
	}
}
