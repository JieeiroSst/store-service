package shared

import "time"

type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

type BaseEvent struct {
	Type  string
	AggID string
	At    time.Time
}

func NewBaseEvent(eventType, aggregateID string, at time.Time) BaseEvent {
	return BaseEvent{Type: eventType, AggID: aggregateID, At: at}
}

func (e BaseEvent) EventType() string     { return e.Type }
func (e BaseEvent) AggregateID() string   { return e.AggID }
func (e BaseEvent) OccurredAt() time.Time { return e.At }
