package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/pkg/metrics"
	"github.com/JIeeiroSst/notifyhub-service/internal/store/mysql"
	"github.com/JIeeiroSst/notifyhub-service/internal/worker"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type Scheduler struct {
	mu     sync.RWMutex
	s      gocron.Scheduler
	jobs   map[string]gocron.Job // notifyJobID → gocron.Job
	store  *mysql.Store
	pool   *worker.Pool
	logger *zap.Logger
}

func New(store *mysql.Store, pool *worker.Pool, log *zap.Logger) (*Scheduler, error) {
	s, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, fmt.Errorf("create gocron: %w", err)
	}
	return &Scheduler{
		s:      s,
		jobs:   make(map[string]gocron.Job),
		store:  store,
		pool:   pool,
		logger: log,
	}, nil
}

func (sc *Scheduler) Start(ctx context.Context) error {
	jobs, err := sc.store.ListActiveJobs(ctx)
	if err != nil {
		return fmt.Errorf("list active jobs: %w", err)
	}
	for _, j := range jobs {
		if err := sc.Register(j); err != nil {
			sc.logger.Warn("failed to register job at startup",
				zap.String("job_id", j.ID),
				zap.Error(err),
			)
		}
	}
	sc.s.Start()
	sc.logger.Info("scheduler started", zap.Int("jobs_loaded", len(jobs)))
	metrics.JobsScheduled.Set(float64(len(jobs)))
	return nil
}

func (sc *Scheduler) Stop() error {
	return sc.s.Shutdown()
}

func (sc *Scheduler) Register(j *model.NotifyJob) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if old, ok := sc.jobs[j.ID]; ok {
		_ = sc.s.RemoveJob(old.ID())
		delete(sc.jobs, j.ID)
	}

	jobID := j.ID
	task := gocron.NewTask(func() {
		ctx := context.Background()
		fullJob, err := sc.store.GetJob(ctx, jobID)
		if err != nil {
			sc.logger.Error("load job for execution",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			return
		}

		if fullJob.MaxRuns > 0 && fullJob.RunCount >= fullJob.MaxRuns {
			sc.logger.Info("max_runs reached — disabling job",
				zap.String("job_id", jobID),
				zap.Int64("max_runs", fullJob.MaxRuns),
			)
			_ = sc.store.UpdateJobStatus(ctx, jobID, model.JobStatusCompleted)
			sc.unregisterLocked(jobID)
			return
		}

		sc.pool.Submit(worker.Task{Job: fullJob})
	})

	def, err := buildDefinition(j)
	if err != nil {
		return err
	}

	gj, err := sc.s.NewJob(
		def,
		task,
		gocron.WithName(j.Name),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return fmt.Errorf("gocron NewJob [%s]: %w", j.ID, err)
	}

	sc.jobs[j.ID] = gj
	sc.logger.Info("job registered",
		zap.String("job_id", j.ID),
		zap.String("name", j.Name),
		zap.String("schedule", string(j.ScheduleType)),
	)
	metrics.JobsScheduled.Set(float64(len(sc.jobs)))
	return nil
}

func (sc *Scheduler) Unregister(notifyJobID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.unregisterLocked(notifyJobID)
}

func (sc *Scheduler) unregisterLocked(notifyJobID string) {
	if gj, ok := sc.jobs[notifyJobID]; ok {
		_ = sc.s.RemoveJob(gj.ID())
		delete(sc.jobs, notifyJobID)
		metrics.JobsScheduled.Set(float64(len(sc.jobs)))
		sc.logger.Info("job unregistered", zap.String("job_id", notifyJobID))
	}
}

func (sc *Scheduler) ListScheduled() []string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	ids := make([]string, 0, len(sc.jobs))
	for id := range sc.jobs {
		ids = append(ids, id)
	}
	return ids
}

func (sc *Scheduler) TriggerNow(ctx context.Context, jobID string) error {
	j, err := sc.store.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}
	if ok := sc.pool.Submit(worker.Task{Job: j}); !ok {
		return fmt.Errorf("worker queue full — try again later")
	}
	sc.logger.Info("job triggered manually", zap.String("job_id", jobID))
	return nil
}

func buildDefinition(j *model.NotifyJob) (gocron.JobDefinition, error) {
	switch j.ScheduleType {
	case model.ScheduleCron:
		if j.CronExpr == "" {
			return nil, fmt.Errorf("cron_expr required for cron schedule")
		}
		return gocron.CronJob(j.CronExpr, false), nil

	case model.ScheduleOnce:
		if j.RunAt == nil {
			return nil, fmt.Errorf("run_at required for once schedule")
		}
		return gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(*j.RunAt)), nil

	case model.ScheduleInterval:
		if j.IntervalSec <= 0 {
			return nil, fmt.Errorf("interval_sec must be > 0")
		}
		return gocron.DurationJob(time.Duration(j.IntervalSec) * time.Second), nil

	default:
		return nil, fmt.Errorf("unknown schedule_type %q — must be cron|once|interval", j.ScheduleType)
	}
}
