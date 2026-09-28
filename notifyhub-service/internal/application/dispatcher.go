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

type Settings struct {
	RetryMax   int
	RetryDelay time.Duration
}

func DefaultSettings() Settings {
	return Settings{RetryMax: 3, RetryDelay: time.Second}
}

type dispatcher struct {
	jobs      port.JobRepository
	history   port.HistoryRepository
	senders   port.SenderRegistry
	fetcher   port.DataFetcher
	renderer  port.TemplateRenderer
	scheduler port.JobScheduler
	metrics   port.Metrics
	settings  Settings
	log       *zap.Logger
}

func NewDispatcher(
	jobs port.JobRepository,
	history port.HistoryRepository,
	senders port.SenderRegistry,
	fetcher port.DataFetcher,
	renderer port.TemplateRenderer,
	scheduler port.JobScheduler,
	metrics port.Metrics,
	settings Settings,
	log *zap.Logger,
) port.DispatchUsecase {
	return &dispatcher{
		jobs:      jobs,
		history:   history,
		senders:   senders,
		fetcher:   fetcher,
		renderer:  renderer,
		scheduler: scheduler,
		metrics:   metrics,
		settings:  settings,
		log:       log,
	}
}

func (d *dispatcher) Execute(ctx context.Context, t port.Task) error {
	job, err := d.jobs.Get(ctx, t.JobID)
	if errors.Is(err, port.ErrNotFound) {
		d.scheduler.Unregister(t.JobID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("load job %s: %w", t.JobID, err)
	}
	log := d.log.With(zap.String("job_id", job.ID), zap.String("job_name", job.Name), zap.Bool("manual", t.Manual))

	if !t.Manual {
		// The job may have been paused or finished after it was queued.
		if job.Status != model.JobStatusActive {
			d.scheduler.Unregister(job.ID)
			return nil
		}
		if job.MaxRunsReached() {
			log.Info("max_runs reached, completing job", zap.Int64("max_runs", job.MaxRuns))
			d.complete(ctx, job.ID)
			return nil
		}
	}

	log.Info("executing job")
	runErr := d.run(ctx, job, log)
	failed := runErr != nil

	if err := d.jobs.RecordRun(ctx, job.ID, d.scheduler.NextRun(job.ID), failed); err != nil {
		log.Warn("record job run", zap.Error(err))
	}
	status := "success"
	if failed {
		status = "fail"
	}
	d.metrics.JobExecuted(job.ScheduleType, status)

	if !t.Manual && (job.ScheduleType == model.ScheduleOnce || (job.MaxRuns > 0 && job.RunCount+1 >= job.MaxRuns)) {
		d.complete(ctx, job.ID)
	}
	return runErr
}

func (d *dispatcher) complete(ctx context.Context, jobID string) {
	d.scheduler.Unregister(jobID)
	if err := d.jobs.UpdateStatus(ctx, jobID, model.JobStatusCompleted); err != nil {
		d.log.Warn("mark job completed", zap.String("job_id", jobID), zap.Error(err))
	}
}

func (d *dispatcher) run(ctx context.Context, job *model.NotifyJob, log *zap.Logger) error {
	switch {
	case job.Channel == nil:
		return fmt.Errorf("channel %s not found", job.ChannelID)
	case !job.Channel.IsActive:
		return fmt.Errorf("channel %s is inactive", job.Channel.Name)
	case job.Template == nil:
		return fmt.Errorf("template %s not found", job.TemplateID)
	}

	sender, err := d.senders.Get(job.Channel.Type)
	if err != nil {
		return err
	}

	data := d.templateData(ctx, job, log)
	subject, body, err := d.renderer.Render(job.Template, data)
	if err != nil {
		log.Error("render failed, aborting job", zap.Error(err))
		return err
	}

	var failures int
	for _, recipient := range job.Recipients {
		if err := d.deliver(ctx, job, sender, port.Message{Recipient: recipient, Subject: subject, Body: body}, log); err != nil {
			failures++
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d of %d notification(s) failed", failures, len(job.Recipients))
	}
	return nil
}

func (d *dispatcher) templateData(ctx context.Context, job *model.NotifyJob, log *zap.Logger) map[string]interface{} {
	data := make(map[string]interface{}, len(job.StaticPayload)+1)
	for k, v := range job.StaticPayload {
		data[k] = v
	}
	if job.DataSource == nil || !job.DataSource.IsActive {
		return data
	}

	start := time.Now()
	fetched, err := d.fetcher.Fetch(ctx, job.DataSource)
	if err != nil {
		d.metrics.FetchObserved(job.DataSource.Name, "error", time.Since(start))
		log.Warn("data fetch failed, using static payload only",
			zap.String("source", job.DataSource.Name), zap.Error(err))
		return data
	}
	d.metrics.FetchObserved(job.DataSource.Name, "ok", time.Since(start))

	if m, ok := fetched.(map[string]interface{}); ok {
		for k, v := range m {
			data[k] = v
		}
	} else if fetched != nil {
		data["data"] = fetched
	}
	return data
}

func (d *dispatcher) deliver(ctx context.Context, job *model.NotifyJob, sender port.Sender, msg port.Message, log *zap.Logger) error {
	ct := job.Channel.Type
	hist := &model.NotifyHistory{
		ID:          uuid.NewString(),
		JobID:       job.ID,
		ChannelType: ct,
		Recipient:   msg.Recipient,
		Subject:     msg.Subject,
		Body:        msg.Body,
		Status:      model.NotifyStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	histOK := true
	if err := d.history.Create(ctx, hist); err != nil {
		histOK = false
		log.Warn("persist history failed", zap.Error(err))
	}

	retries, sendErr := d.sendWithRetry(ctx, sender, msg)

	status, errMsg, metric := model.NotifyStatusSent, "", "sent"
	if sendErr != nil {
		status, errMsg, metric = model.NotifyStatusFailed, sendErr.Error(), "failed"
		log.Error("send failed", zap.String("recipient", msg.Recipient), zap.String("channel", string(ct)), zap.Error(sendErr))
	} else {
		log.Info("notification sent", zap.String("recipient", msg.Recipient), zap.String("channel", string(ct)))
	}
	if histOK {
		if err := d.history.UpdateStatus(ctx, hist.ID, status, retries, errMsg); err != nil {
			log.Warn("update history failed", zap.Error(err))
		}
	}
	d.metrics.NotificationSent(ct, metric)
	return sendErr
}

func (d *dispatcher) sendWithRetry(ctx context.Context, sender port.Sender, msg port.Message) (int, error) {
	var lastErr error
	for attempt := 0; attempt <= d.settings.RetryMax; attempt++ {
		if attempt > 0 {
			delay := d.settings.RetryDelay * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return attempt - 1, ctx.Err()
			case <-time.After(delay):
			}
			d.metrics.SendRetried(sender.Type())
		}
		if lastErr = sender.Send(ctx, msg); lastErr == nil {
			return attempt, nil
		}
	}
	return d.settings.RetryMax, fmt.Errorf("after %d retries: %w", d.settings.RetryMax, lastErr)
}
