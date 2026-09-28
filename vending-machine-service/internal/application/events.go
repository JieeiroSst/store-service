package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
)

func newEvent(at time.Time, eventType, entity, entityID, machineID string, data map[string]any) *domain.Event {
	if data == nil {
		data = map[string]any{}
	}
	return &domain.Event{
		ID:            uuid.NewString(),
		EventType:     eventType,
		RelatedEntity: entity,
		EntityID:      entityID,
		MachineID:     machineID,
		Data:          data,
		OccurredAt:    at,
	}
}

func appendEvents(ctx context.Context, repo port.EventRepository, events ...*domain.Event) error {
	for _, e := range events {
		if err := repo.Append(ctx, e); err != nil {
			return err
		}
	}
	return nil
}
