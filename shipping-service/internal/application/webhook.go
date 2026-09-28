package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type webhookService struct {
	lifecycle *lifecycle
	shipments port.ShipmentRepository
	carrier   port.Carrier
}

func NewWebhookService(
	shipments port.ShipmentRepository,
	events port.EventRepository,
	outbox port.OutboxRepository,
	carrier port.Carrier,
	tx port.TxManager,
	clock port.Clock,
) port.WebhookUsecase {
	return &webhookService{
		lifecycle: &lifecycle{shipments: shipments, events: events, outbox: outbox, tx: tx, clock: clock},
		shipments: shipments,
		carrier:   carrier,
	}
}

func (s *webhookService) HandleGHN(ctx context.Context, token string, ev port.CarrierWebhook) error {
	if !s.carrier.VerifyWebhookToken(token) {
		return port.ErrUnauthorized
	}
	kind := strings.ToLower(strings.TrimSpace(ev.Type))
	if kind != "create" && kind != "switch_status" {
		return nil
	}
	status, ok := model.MapGHNStatus(strings.ToLower(strings.TrimSpace(ev.Status)))
	if !ok {
		return nil
	}

	shipment, err := s.find(ctx, ev)
	if errors.Is(err, port.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	upd := statusUpdate{
		To:            status,
		CarrierStatus: strings.ToLower(ev.Status),
		Reason:        ev.Reason,
		Source:        model.SourceWebhook,
	}
	if t, err := time.Parse(time.RFC3339, ev.Time); err == nil {
		upd.OccurredAt = t
	}
	if ev.OrderCode != "" && shipment.CarrierOrderCode == "" {
		code := ev.OrderCode
		upd.Mutate = func(sh *model.Shipment) { sh.CarrierOrderCode = code }
	}
	_, _, err = s.lifecycle.apply(ctx, shipment.ID, upd)
	return err
}

func (s *webhookService) find(ctx context.Context, ev port.CarrierWebhook) (*model.Shipment, error) {
	if ev.OrderCode != "" {
		shipment, err := s.shipments.GetByCarrierOrderCode(ctx, ev.OrderCode)
		if !errors.Is(err, port.ErrNotFound) {
			return shipment, err
		}
	}
	if ev.ClientOrderCode != "" {
		return s.shipments.GetByCode(ctx, ev.ClientOrderCode)
	}
	return nil, port.ErrNotFound
}
