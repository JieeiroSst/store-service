package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type lifecycle struct {
	shipments port.ShipmentRepository
	events    port.EventRepository
	outbox    port.OutboxRepository
	tx        port.TxManager
	clock     port.Clock
}

type statusUpdate struct {
	To            model.Status
	CarrierStatus string
	Reason        string
	Source        string
	OccurredAt    time.Time
	Mutate        func(s *model.Shipment)
}

func (l *lifecycle) apply(ctx context.Context, shipmentID int64, upd statusUpdate) (*model.Shipment, bool, error) {
	if upd.OccurredAt.IsZero() {
		upd.OccurredAt = l.clock.Now()
	}
	var (
		result  *model.Shipment
		applied bool
	)
	err := l.tx.WithinTx(ctx, func(ctx context.Context) error {
		s, err := l.shipments.GetByIDForUpdate(ctx, shipmentID)
		if err != nil {
			return err
		}
		result = s
		from := s.Status
		applied = from != upd.To && model.CanTransition(from, upd.To)

		event := &model.ShipmentEvent{
			ShipmentID:    s.ID,
			Source:        upd.Source,
			FromStatus:    from,
			ToStatus:      upd.To,
			CarrierStatus: upd.CarrierStatus,
			Reason:        upd.Reason,
			Applied:       applied,
			OccurredAt:    upd.OccurredAt.UTC().Truncate(time.Millisecond),
		}
		inserted, err := l.events.Create(ctx, event)
		if err != nil {
			return err
		}
		if !inserted {
			applied = false
			return nil
		}

		if !applied && from != upd.To {
			return nil
		}
		if upd.Mutate != nil {
			upd.Mutate(s)
		}
		if upd.CarrierStatus != "" {
			s.CarrierStatus = upd.CarrierStatus
		}
		if applied {
			s.Status = upd.To
			if upd.To == model.StatusDelivered {
				t := event.OccurredAt
				s.DeliveredAt = &t
			}
			jobs, err := l.jobsFor(*s, from, *event)
			if err != nil {
				return err
			}
			if err := l.outbox.Enqueue(ctx, jobs); err != nil {
				return err
			}
		}
		return l.shipments.Save(ctx, s)
	})
	if err != nil {
		return nil, false, err
	}
	return result, applied, nil
}

func (l *lifecycle) jobsFor(s model.Shipment, from model.Status, event model.ShipmentEvent) ([]model.OutboxJob, error) {
	now := l.clock.Now()
	var jobs []model.OutboxJob
	for _, n := range model.CustomerNotifications(s) {
		payload, err := json.Marshal(n)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, model.OutboxJob{Kind: model.JobNotify, ShipmentID: s.ID, Payload: string(payload), State: model.JobPending, NextAttemptAt: now})
	}
	if s.CallbackURL != "" {
		payload, err := json.Marshal(model.CallbackJob{URL: s.CallbackURL, Event: model.CallbackEvent{
			EventID:         fmt.Sprintf("evt-%d", event.ID),
			ShipmentID:      s.ID,
			ShipmentCode:    s.Code,
			ClientOrderCode: s.ClientOrderCode,
			Status:          s.Status,
			PreviousStatus:  from,
			CarrierStatus:   s.CarrierStatus,
			CarrierOrder:    s.CarrierOrderCode,
			Reason:          event.Reason,
			OccurredAt:      event.OccurredAt,
		}})
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, model.OutboxJob{Kind: model.JobCallback, ShipmentID: s.ID, Payload: string(payload), State: model.JobPending, NextAttemptAt: now})
	}
	return jobs, nil
}
