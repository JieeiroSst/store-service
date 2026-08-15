package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Event struct {
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	Topic         string
	OccurredAt    time.Time
}

func NewEventFromDomain(aggregateType, topic string, evt shared.DomainEvent) (Event, error) {
	payload, err := json.Marshal(evt)
	if err != nil {
		return Event{}, err
	}
	return Event{
		AggregateType: aggregateType,
		AggregateID:   evt.AggregateID(),
		EventType:     evt.EventType(),
		Payload:       payload,
		Topic:         topic,
		OccurredAt:    evt.OccurredAt(),
	}, nil
}
type Outbox interface {
	Enqueue(ctx context.Context, event Event) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
}

type StoredEvent struct {
	ID            string
	AggregateID   string
	EventType     string
	Payload       []byte
	Topic         string
	Attempts      int
}

type Repository interface {
	Outbox
	FetchUnpublished(ctx context.Context, limit int) ([]StoredEvent, error)
	MarkPublished(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, lastErr string) error
}
