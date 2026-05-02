package job

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/pkg/metrics"
	"github.com/JIeeiroSst/notifyhub-service/internal/scheduler"
	"github.com/JIeeiroSst/notifyhub-service/internal/store/mysql"
	tmplEngine "github.com/JIeeiroSst/notifyhub-service/internal/template"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	store  *mysql.Store
	sched  *scheduler.Scheduler
	tmpl   *tmplEngine.Engine
	logger *zap.Logger
}

func NewService(
	store *mysql.Store,
	sched *scheduler.Scheduler,
	tmpl *tmplEngine.Engine,
	log *zap.Logger,
) *Service {
	return &Service{store: store, sched: sched, tmpl: tmpl, logger: log}
}

func (s *Service) CreateJob(ctx context.Context, j *model.NotifyJob) (*model.NotifyJob, error) {
	if err := validateJob(j); err != nil {
		return nil, err
	}
	j.ID = uuid.NewString()
	j.Status = model.JobStatusActive
	j.CreatedAt = time.Now()
	j.UpdatedAt = time.Now()

	if err := s.store.CreateJob(ctx, j); err != nil {
		return nil, fmt.Errorf("persist job: %w", err)
	}

	full, err := s.store.GetJob(ctx, j.ID)
	if err != nil {
		return nil, fmt.Errorf("reload job: %w", err)
	}

	if err := s.sched.Register(full); err != nil {
		s.logger.Warn("scheduler register", zap.String("job_id", j.ID), zap.Error(err))
	}

	metrics.JobsScheduled.Inc()
	return full, nil
}

func (s *Service) UpdateJob(ctx context.Context, id string, patch *model.NotifyJob) (*model.NotifyJob, error) {
	existing, err := s.store.GetJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	patch.ID = existing.ID
	patch.CreatedAt = existing.CreatedAt
	patch.UpdatedAt = time.Now()

	if err := validateJob(patch); err != nil {
		return nil, err
	}
	if err := s.store.UpdateJob(ctx, patch); err != nil {
		return nil, fmt.Errorf("update job: %w", err)
	}
	full, _ := s.store.GetJob(ctx, id)
	if full != nil && full.Status == model.JobStatusActive {
		_ = s.sched.Register(full)
	}
	return full, nil
}

func (s *Service) PauseJob(ctx context.Context, id string) error {
	if err := s.store.UpdateJobStatus(ctx, id, model.JobStatusPaused); err != nil {
		return err
	}
	s.sched.Unregister(id)
	metrics.JobsScheduled.Dec()
	return nil
}

func (s *Service) ResumeJob(ctx context.Context, id string) error {
	if err := s.store.UpdateJobStatus(ctx, id, model.JobStatusActive); err != nil {
		return err
	}
	j, err := s.store.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if err := s.sched.Register(j); err != nil {
		return fmt.Errorf("re-register scheduler: %w", err)
	}
	metrics.JobsScheduled.Inc()
	return nil
}

func (s *Service) DeleteJob(ctx context.Context, id string) error {
	s.sched.Unregister(id)
	metrics.JobsScheduled.Dec()
	return s.store.DeleteJob(ctx, id)
}

func (s *Service) TriggerNow(ctx context.Context, id string) error {
	return s.sched.TriggerNow(ctx, id)
}

func validateJob(j *model.NotifyJob) error {
	if j.Name == "" {
		return fmt.Errorf("job name is required")
	}
	if j.ChannelID == "" {
		return fmt.Errorf("channel_id is required")
	}
	if j.TemplateID == "" {
		return fmt.Errorf("template_id is required")
	}
	if len(j.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	switch j.ScheduleType {
	case model.ScheduleCron:
		if j.CronExpr == "" {
			return fmt.Errorf("cron_expr required for cron schedule")
		}
	case model.ScheduleOnce:
		if j.RunAt == nil {
			return fmt.Errorf("run_at required for once schedule")
		}
		if j.RunAt.Before(time.Now()) {
			return fmt.Errorf("run_at must be in the future")
		}
	case model.ScheduleInterval:
		if j.IntervalSec <= 0 {
			return fmt.Errorf("interval_sec must be > 0")
		}
	default:
		return fmt.Errorf("schedule_type must be cron|once|interval")
	}
	return nil
}
