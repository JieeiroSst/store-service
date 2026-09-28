package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type jobService struct {
	jobs      port.JobRepository
	channels  port.ChannelRepository
	templates port.TemplateRepository
	sources   port.DataSourceRepository
	scheduler port.JobScheduler
	queue     port.TaskQueue
	log       *zap.Logger
}

func NewJobService(
	jobs port.JobRepository,
	channels port.ChannelRepository,
	templates port.TemplateRepository,
	sources port.DataSourceRepository,
	scheduler port.JobScheduler,
	queue port.TaskQueue,
	log *zap.Logger,
) port.JobUsecase {
	return &jobService{
		jobs:      jobs,
		channels:  channels,
		templates: templates,
		sources:   sources,
		scheduler: scheduler,
		queue:     queue,
		log:       log,
	}
}

func validateJob(j *model.NotifyJob) error {
	if j.Name == "" {
		return invalid("job name is required")
	}
	if j.ChannelID == "" {
		return invalid("channel_id is required")
	}
	if j.TemplateID == "" {
		return invalid("template_id is required")
	}
	if len(j.Recipients) == 0 {
		return invalid("at least one recipient is required")
	}
	if j.MaxRuns < 0 {
		return invalid("max_runs must be >= 0")
	}
	switch j.ScheduleType {
	case model.ScheduleCron:
		if j.CronExpr == "" {
			return invalid("cron_expr required for cron schedule")
		}
	case model.ScheduleOnce:
		if j.RunAt == nil {
			return invalid("run_at required for once schedule")
		}
		if j.RunAt.Before(time.Now()) {
			return invalid("run_at must be in the future")
		}
	case model.ScheduleInterval:
		if j.IntervalSec <= 0 {
			return invalid("interval_sec must be > 0")
		}
	default:
		return invalid("schedule_type must be cron|once|interval")
	}
	return nil
}

func (s *jobService) checkReferences(ctx context.Context, j *model.NotifyJob) error {
	ch, err := s.channels.Get(ctx, j.ChannelID)
	if err != nil {
		return refErr("channel_id", err)
	}
	t, err := s.templates.Get(ctx, j.TemplateID)
	if err != nil {
		return refErr("template_id", err)
	}
	if t.Channel != ch.Type {
		return invalid("template channel %q does not match channel type %q", t.Channel, ch.Type)
	}
	if j.DataSourceID != "" {
		if _, err := s.sources.Get(ctx, j.DataSourceID); err != nil {
			return refErr("data_source_id", err)
		}
	}
	return nil
}

func refErr(field string, err error) error {
	if errors.Is(err, port.ErrNotFound) {
		return invalid("%s does not exist", field)
	}
	return err
}

func (s *jobService) Create(ctx context.Context, j *model.NotifyJob) (*model.NotifyJob, error) {
	if err := validateJob(j); err != nil {
		return nil, err
	}
	if err := s.checkReferences(ctx, j); err != nil {
		return nil, err
	}
	now := time.Now()
	j.ID = uuid.NewString()
	j.Status = model.JobStatusActive
	j.RunCount, j.FailCount = 0, 0
	j.LastRunAt, j.NextRunAt = nil, nil
	j.CreatedAt, j.UpdatedAt = now, now
	j.Channel, j.Template, j.DataSource = nil, nil, nil

	if err := s.jobs.Create(ctx, j); err != nil {
		return nil, fmt.Errorf("persist job: %w", err)
	}
	if err := s.scheduler.Register(j); err != nil {
		// e.g. an unparsable cron expression: do not keep a job that can never run.
		if delErr := s.jobs.Delete(ctx, j.ID); delErr != nil {
			s.log.Error("rollback job after schedule failure", zap.String("job_id", j.ID), zap.Error(delErr))
		}
		return nil, invalid("schedule: %v", err)
	}
	return s.jobs.Get(ctx, j.ID)
}

func (s *jobService) Get(ctx context.Context, id string) (*model.NotifyJob, error) {
	return s.jobs.Get(ctx, id)
}

func (s *jobService) List(ctx context.Context, f port.JobFilter) ([]*model.NotifyJob, int64, error) {
	return s.jobs.List(ctx, f)
}

func (s *jobService) Update(ctx context.Context, id string, apply func(*model.NotifyJob) error) (*model.NotifyJob, error) {
	j, err := s.jobs.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	keep := *j
	if err := apply(j); err != nil {
		return nil, invalid("%v", err)
	}
	j.ID, j.CreatedAt, j.UpdatedAt = id, keep.CreatedAt, time.Now()
	j.Status, j.RunCount, j.FailCount = keep.Status, keep.RunCount, keep.FailCount
	j.LastRunAt, j.NextRunAt = keep.LastRunAt, keep.NextRunAt
	j.Channel, j.Template, j.DataSource = nil, nil, nil

	if err := validateJob(j); err != nil {
		return nil, err
	}
	if err := s.checkReferences(ctx, j); err != nil {
		return nil, err
	}
	if j.Status == model.JobStatusActive {
		if err := s.scheduler.Register(j); err != nil {
			return nil, invalid("schedule: %v", err)
		}
	}
	if err := s.jobs.Update(ctx, j); err != nil {
		// Put the previous schedule back so memory matches the database.
		if keep.Status == model.JobStatusActive {
			_ = s.scheduler.Register(&keep)
		}
		return nil, fmt.Errorf("update job: %w", err)
	}
	return s.jobs.Get(ctx, id)
}

func (s *jobService) Pause(ctx context.Context, id string) error {
	if err := s.jobs.UpdateStatus(ctx, id, model.JobStatusPaused); err != nil {
		return err
	}
	s.scheduler.Unregister(id)
	return nil
}

func (s *jobService) Resume(ctx context.Context, id string) error {
	j, err := s.jobs.Get(ctx, id)
	if err != nil {
		return err
	}
	if j.ScheduleType == model.ScheduleOnce && j.RunAt != nil && j.RunAt.Before(time.Now()) {
		return invalid("run_at is in the past; update the job with a new run_at first")
	}
	if err := s.scheduler.Register(j); err != nil {
		return invalid("schedule: %v", err)
	}
	if err := s.jobs.UpdateStatus(ctx, id, model.JobStatusActive); err != nil {
		s.scheduler.Unregister(id)
		return err
	}
	return nil
}

func (s *jobService) Delete(ctx context.Context, id string) error {
	if err := s.jobs.Delete(ctx, id); err != nil {
		return err
	}
	s.scheduler.Unregister(id)
	return nil
}

func (s *jobService) Trigger(ctx context.Context, id string) error {
	if _, err := s.jobs.Get(ctx, id); err != nil {
		return err
	}
	return s.queue.Enqueue(port.Task{JobID: id, Manual: true})
}

func (s *jobService) RestoreSchedules(ctx context.Context) error {
	jobs, err := s.jobs.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active jobs: %w", err)
	}
	registered := 0
	for _, j := range jobs {
		if j.ScheduleType == model.ScheduleOnce && j.RunAt != nil && !j.RunAt.After(time.Now()) {
			s.restoreMissedOnce(ctx, j)
			continue
		}
		if err := s.scheduler.Register(j); err != nil {
			s.log.Warn("failed to register job at startup", zap.String("job_id", j.ID), zap.Error(err))
			continue
		}
		registered++
	}
	s.log.Info("schedules restored", zap.Int("active", len(jobs)), zap.Int("registered", registered))
	return nil
}

func (s *jobService) restoreMissedOnce(ctx context.Context, j *model.NotifyJob) {
	if j.LastRunAt != nil {
		if err := s.jobs.UpdateStatus(ctx, j.ID, model.JobStatusCompleted); err != nil {
			s.log.Warn("complete once job", zap.String("job_id", j.ID), zap.Error(err))
		}
		return
	}
	s.log.Info("once job missed its run_at, running now", zap.String("job_id", j.ID))
	if err := s.queue.Enqueue(port.Task{JobID: j.ID}); err != nil {
		s.log.Warn("enqueue missed once job", zap.String("job_id", j.ID), zap.Error(err))
	}
}

func (s *jobService) ScheduledJobs() []string {
	return s.scheduler.Scheduled()
}
