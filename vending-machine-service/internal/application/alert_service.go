package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

const alertBatch = 50

type alertService struct {
	tx       port.Transactor
	events   port.EventRepository
	notifier port.Notifier
	now      func() time.Time
}

func NewAlertService(tx port.Transactor, events port.EventRepository, notifier port.Notifier) port.AlertService {
	return &alertService{tx: tx, events: events, notifier: notifier, now: time.Now}
}

func (s *alertService) DispatchPending(ctx context.Context) (int, error) {
	var (
		sent      int
		notifyErr error
	)
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		events, err := s.events.LockUndispatched(ctx, alertBatch)
		if err != nil {
			return err
		}
		done := make([]string, 0, len(events))
		for _, e := range events {
			if e.IsAlert() {
				if notifyErr = s.notifier.Notify(ctx, e); notifyErr != nil {
					break
				}
				sent++
			}
			done = append(done, e.ID)
		}
		if len(done) == 0 {
			return nil
		}
		return s.events.MarkDispatched(ctx, done, s.now().UTC())
	})
	if err != nil {
		return 0, err
	}
	return sent, notifyErr
}
