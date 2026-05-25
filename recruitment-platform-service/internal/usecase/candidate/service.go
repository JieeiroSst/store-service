package candidateusecase

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type service struct {
	repo     candidate.Repository
	aiSvc    port.AIService
	eventBus port.EventBus
	logger   *zap.Logger
}

func NewService(
	repo candidate.Repository,
	aiSvc port.AIService,
	eventBus port.EventBus,
	logger *zap.Logger,
) port.CandidateService {
	return &service{
		repo:     repo,
		aiSvc:    aiSvc,
		eventBus: eventBus,
		logger:   logger,
	}
}

func (s *service) Create(ctx context.Context, cmd port.CreateCandidateCommand) (*candidate.Candidate, error) {
	// Dedup by email
	existing, _ := s.repo.FindByEmail(ctx, cmd.Email)
	if existing != nil {
		return nil, errors.New("candidate: email already registered")
	}

	c, err := candidate.New(cmd.FullName, cmd.Email, cmd.Phone, cmd.Source)
	if err != nil {
		return nil, err
	}
	c.ResumeURL = cmd.ResumeURL

	if err := s.repo.Save(ctx, c); err != nil {
		return nil, err
	}

	if cmd.ResumeURL != "" {
		go func() {
			if err := s.EnrichWithAI(context.Background(), c.ID); err != nil {
				s.logger.Warn("resume enrichment failed",
					zap.String("candidate_id", c.ID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	s.publishEvents(ctx, c.DomainEvents())
	c.ClearEvents()
	return c, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*candidate.Candidate, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if v, ok := updates["full_name"].(string); ok {
		c.FullName = v
	}
	if v, ok := updates["phone"].(string); ok {
		c.Phone = v
	}
	if v, ok := updates["current_title"].(string); ok {
		c.CurrentTitle = v
	}
	if v, ok := updates["current_company"].(string); ok {
		c.CurrentCompany = v
	}
	if v, ok := updates["linkedin_url"].(string); ok {
		c.LinkedInURL = v
	}
	if v, ok := updates["resume_url"].(string); ok {
		c.ResumeURL = v
	}
	if v, ok := updates["notes"].(string); ok {
		c.Notes = v
	}
	if v, ok := updates["notice_period_days"].(int); ok {
		c.NoticePeriodDays = v
	}
	if v, ok := updates["skills"].([]string); ok {
		c.Skills = v
	}
	if v, ok := updates["tags"].([]string); ok {
		c.Tags = v
	}

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *service) Transition(ctx context.Context, id uuid.UUID, status candidate.Status, reason string) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := c.TransitionTo(status, reason); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return err
	}
	s.publishEvents(ctx, c.DomainEvents())
	c.ClearEvents()
	return nil
}

func (s *service) EnrichWithAI(ctx context.Context, id uuid.UUID) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if c.ResumeURL == "" {
		return errors.New("candidate: no resume URL to parse")
	}

	parsed, err := s.aiSvc.ParseResume(ctx, c.ResumeURL)
	if err != nil {
		return err
	}

	c.Skills = parsed.Skills
	c.YearsOfExperience = parsed.YearsExperience
	c.ExperienceLevel = parsed.ExperienceLevel
	c.Embedding = parsed.Embedding
	if parsed.SuggestedTitle != "" && c.CurrentTitle == "" {
		c.CurrentTitle = parsed.SuggestedTitle
	}

	return s.repo.Update(ctx, c)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*candidate.Candidate, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter candidate.Filter) (shared.PaginatedResult[*candidate.Candidate], error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *service) publishEvents(ctx context.Context, events []shared.DomainEvent) {
	for _, e := range events {
		if err := s.eventBus.Publish(ctx, e); err != nil {
			s.logger.Error("event publish failed",
				zap.String("type", e.EventType),
				zap.Error(err),
			)
		}
	}
}
