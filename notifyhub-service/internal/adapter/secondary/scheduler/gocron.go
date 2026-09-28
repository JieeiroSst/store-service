package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type Scheduler struct {
	mu    sync.Mutex
	s     gocron.Scheduler
	jobs  map[string]gocron.Job
	queue port.TaskQueue
	log   *zap.Logger
}

func New(queue port.TaskQueue, log *zap.Logger) (*Scheduler, error) {
	s, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, fmt.Errorf("create gocron: %w", err)
	}
	return &Scheduler{s: s, jobs: make(map[string]gocron.Job), queue: queue, log: log}, nil
}

func (sc *Scheduler) Start() { sc.s.Start() }

func (sc *Scheduler) Shutdown() error { return sc.s.Shutdown() }

func (sc *Scheduler) Register(j *model.NotifyJob) error {
	def, err := definition(j)
	if err != nil {
		return err
	}

	jobID := j.ID
	task := gocron.NewTask(func() {
		if err := sc.queue.Enqueue(port.Task{JobID: jobID}); err != nil {
			sc.log.Warn("scheduled run dropped", zap.String("job_id", jobID), zap.Error(err))
		}
	})

	sc.mu.Lock()
	defer sc.mu.Unlock()

	gj, err := sc.s.NewJob(def, task,
		gocron.WithName(j.Name),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return fmt.Errorf("schedule job %s: %w", j.ID, err)
	}
	if old, ok := sc.jobs[jobID]; ok {
		_ = sc.s.RemoveJob(old.ID())
	}
	sc.jobs[jobID] = gj
	metrics.JobsScheduled.Set(float64(len(sc.jobs)))
	sc.log.Info("job registered",
		zap.String("job_id", jobID),
		zap.String("name", j.Name),
		zap.String("schedule", string(j.ScheduleType)),
	)
	return nil
}

func (sc *Scheduler) Unregister(jobID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	gj, ok := sc.jobs[jobID]
	if !ok {
		return
	}
	_ = sc.s.RemoveJob(gj.ID())
	delete(sc.jobs, jobID)
	metrics.JobsScheduled.Set(float64(len(sc.jobs)))
	sc.log.Info("job unregistered", zap.String("job_id", jobID))
}

func (sc *Scheduler) Scheduled() []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	ids := make([]string, 0, len(sc.jobs))
	for id := range sc.jobs {
		ids = append(ids, id)
	}
	return ids
}

func (sc *Scheduler) NextRun(jobID string) *time.Time {
	sc.mu.Lock()
	gj, ok := sc.jobs[jobID]
	sc.mu.Unlock()
	if !ok {
		return nil
	}
	next, err := gj.NextRun()
	if err != nil || next.IsZero() {
		return nil
	}
	return &next
}

func definition(j *model.NotifyJob) (gocron.JobDefinition, error) {
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
		return nil, fmt.Errorf("unknown schedule_type %q, must be cron|once|interval", j.ScheduleType)
	}
}
