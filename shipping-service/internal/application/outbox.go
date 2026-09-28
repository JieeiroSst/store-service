package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

const (
	retryBase = 5 * time.Second
	retryCap  = time.Hour
)

type outboxService struct {
	outbox    port.OutboxRepository
	notifier  port.Notifier
	callbacks port.CallbackSender
	clock     port.Clock
	settings  Settings
}

func NewOutboxService(outbox port.OutboxRepository, notifier port.Notifier, callbacks port.CallbackSender, clock port.Clock, settings Settings) port.OutboxUsecase {
	return &outboxService{outbox: outbox, notifier: notifier, callbacks: callbacks, clock: clock, settings: settings}
}

func (s *outboxService) DispatchDue(ctx context.Context) (int, error) {
	jobs, err := s.outbox.ClaimDue(ctx, s.clock.Now(), s.settings.OutboxBatch, s.settings.OutboxLease)
	if err != nil {
		return 0, err
	}
	for _, job := range jobs {
		if err := s.deliver(ctx, job); err != nil {
			if markErr := s.fail(ctx, job, err); markErr != nil {
				return 0, markErr
			}
			continue
		}
		if err := s.outbox.MarkDone(ctx, job.ID); err != nil {
			return 0, err
		}
	}
	return len(jobs), nil
}

func (s *outboxService) deliver(ctx context.Context, job model.OutboxJob) error {
	switch job.Kind {
	case model.JobNotify:
		var n model.Notification
		if err := json.Unmarshal([]byte(job.Payload), &n); err != nil {
			return err
		}
		return s.notifier.Notify(ctx, n)
	case model.JobCallback:
		var c model.CallbackJob
		if err := json.Unmarshal([]byte(job.Payload), &c); err != nil {
			return err
		}
		return s.callbacks.Send(ctx, c.URL, c.Event)
	default:
		return fmt.Errorf("unknown job kind %q", job.Kind)
	}
}

func (s *outboxService) fail(ctx context.Context, job model.OutboxJob, cause error) error {
	if job.Attempts >= s.settings.OutboxMaxAttempts {
		return s.outbox.MarkFailed(ctx, job.ID, cause.Error())
	}
	return s.outbox.MarkRetry(ctx, job.ID, s.clock.Now().Add(Backoff(job.Attempts)), cause.Error())
}

func Backoff(attempts int) time.Duration {
	d := retryBase
	for i := 1; i < attempts && d < retryCap; i++ {
		d *= 2
	}
	return min(d, retryCap)
}
