package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/channel"
	"github.com/JIeeiroSst/notifyhub-service/internal/fetcher"
	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/pkg/metrics"
	"github.com/JIeeiroSst/notifyhub-service/internal/store"
	tmplEngine "github.com/JIeeiroSst/notifyhub-service/internal/template"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Task struct {
	Job *model.NotifyJob
}

type Pool struct {
	queue    chan Task
	wg       sync.WaitGroup
	logger   *zap.Logger
	registry *channel.Registry
	store    store.JobStore
	fetcher  *fetcher.Fetcher
	tmpl     *tmplEngine.Engine
	retryMax int
	retryMs  int
	size     int
}

func NewPool(
	size, queueSize, retryMax, retryMs int,
	registry *channel.Registry,
	st store.JobStore,
	f *fetcher.Fetcher,
	tmpl *tmplEngine.Engine,
	log *zap.Logger,
) *Pool {
	p := &Pool{
		queue:    make(chan Task, queueSize),
		logger:   log,
		registry: registry,
		store:    st,
		fetcher:  f,
		tmpl:     tmpl,
		retryMax: retryMax,
		retryMs:  retryMs,
		size:     size,
	}
	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	return p
}

func (p *Pool) Submit(t Task) bool {
	depth := float64(len(p.queue))
	metrics.WorkerQueueDepth.Set(depth)

	select {
	case p.queue <- t:
		return true
	default:
		p.logger.Warn("worker queue full — dropping task",
			zap.String("job_id", t.Job.ID),
			zap.Int("queue_cap", cap(p.queue)),
		)
		return false
	}
}

func (p *Pool) Stop() {
	close(p.queue)
	p.wg.Wait()
}

func (p *Pool) QueueDepth() int { return len(p.queue) }

func (p *Pool) WorkerCount() int { return p.size }


func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for task := range p.queue {
		metrics.WorkerQueueDepth.Set(float64(len(p.queue)))
		p.process(context.Background(), task, id)
	}
}

func (p *Pool) process(ctx context.Context, task Task, workerID int) {
	job := task.Job
	log := p.logger.With(
		zap.String("job_id", job.ID),
		zap.String("job_name", job.Name),
		zap.Int("worker", workerID),
	)
	log.Info("executing job")

	templateData := make(map[string]interface{})
	for k, v := range job.StaticPayload {
		templateData[k] = v
	}

	if job.DataSource != nil && job.DataSource.IsActive {
		fetchStart := time.Now()
		fetched, err := p.fetcher.Fetch(ctx, job.DataSource)
		fetchLatency := time.Since(fetchStart).Seconds()

		if err != nil {
			log.Warn("data fetch failed — using static payload only",
				zap.String("source", job.DataSource.Name),
				zap.Error(err),
			)
			metrics.FetchDuration.WithLabelValues(job.DataSource.Name, "error").Observe(fetchLatency)
		} else {
			metrics.FetchDuration.WithLabelValues(job.DataSource.Name, "ok").Observe(fetchLatency)
			mergeData(templateData, fetched)
			log.Debug("data fetched", zap.String("source", job.DataSource.Name))
		}
	}

	subject, err := p.tmpl.RenderSubject(job.Template.Subject, templateData)
	if err != nil {
		log.Warn("render subject error", zap.Error(err))
		subject = job.Template.Subject // fall back to raw
	}

	body, err := p.tmpl.Render(ctx, job.TemplateID, templateData)
	if err != nil {
		log.Error("render body failed — aborting job", zap.Error(err))
		_ = p.store.UpdateJobRun(ctx, job.ID, nil, true)
		metrics.JobExecutionsTotal.WithLabelValues(string(job.ScheduleType), "fail").Inc()
		return
	}

	var anyFailed bool
	for _, recipient := range job.Recipients {
		msg := channel.Message{
			Recipient: recipient,
			Subject:   subject,
			Body:      body,
		}

		// Record history entry (pending)
		histID := uuid.NewString()
		hist := &model.NotifyHistory{
			ID:          histID,
			JobID:       job.ID,
			ChannelType: job.Channel.Type,
			Recipient:   recipient,
			Subject:     subject,
			Body:        body,
			Status:      model.NotifyStatusPending,
			CreatedAt:   time.Now(),
		}
		if dbErr := p.store.CreateHistory(ctx, hist); dbErr != nil {
			log.Warn("persist history failed", zap.Error(dbErr))
		}

		sendErr := p.sendWithRetry(ctx, job.Channel.Type, msg)
		if sendErr != nil {
			anyFailed = true
			log.Error("send failed",
				zap.String("recipient", recipient),
				zap.String("channel", string(job.Channel.Type)),
				zap.Error(sendErr),
			)
			_ = p.store.UpdateHistoryStatus(ctx, histID, model.NotifyStatusFailed, sendErr.Error())
			metrics.NotificationsSentTotal.WithLabelValues(string(job.Channel.Type), "failed").Inc()
		} else {
			log.Info("notification sent",
				zap.String("recipient", recipient),
				zap.String("channel", string(job.Channel.Type)),
			)
			_ = p.store.UpdateHistoryStatus(ctx, histID, model.NotifyStatusSent, "")
			metrics.NotificationsSentTotal.WithLabelValues(string(job.Channel.Type), "sent").Inc()
		}
	}

	_ = p.store.UpdateJobRun(ctx, job.ID, nil, anyFailed)
	if anyFailed {
		metrics.JobExecutionsTotal.WithLabelValues(string(job.ScheduleType), "fail").Inc()
	} else {
		metrics.JobExecutionsTotal.WithLabelValues(string(job.ScheduleType), "success").Inc()
	}
}

func (p *Pool) sendWithRetry(ctx context.Context, ct model.ChannelType, msg channel.Message) error {
	sender, err := p.registry.Get(ct)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= p.retryMax; attempt++ {
		if attempt > 0 {
			delay := time.Duration(p.retryMs*(1<<(attempt-1))) * time.Millisecond
			p.logger.Debug("retry send",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
				zap.String("recipient", msg.Recipient),
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			metrics.RetryTotal.WithLabelValues(string(ct)).Inc()
		}

		lastErr = sender.Send(ctx, msg)
		if lastErr == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d retries: %w", p.retryMax, lastErr)
}

func mergeData(dest map[string]interface{}, src interface{}) {
	if src == nil {
		return
	}
	switch v := src.(type) {
	case map[string]interface{}:
		for k, val := range v {
			dest[k] = val
		}
	default:
		b, _ := json.Marshal(src)
		var arr interface{}
		if json.Unmarshal(b, &arr) == nil {
			dest["data"] = arr
		}
	}
}
