package shared

import (
	"time"

	"github.com/google/uuid"
)

type BaseEntity struct {
	ID        uuid.UUID  `db:"id"        json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

func NewBaseEntity() BaseEntity {
	now := time.Now()
	return BaseEntity{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type PaginationParams struct {
	Page  int `form:"page"  json:"page"`
	Limit int `form:"limit" json:"limit"`
}

func (p *PaginationParams) Offset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 20
	}
	return (p.Page - 1) * p.Limit
}

type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

type DomainEvent struct {
	EventID   uuid.UUID   `json:"event_id"`
	EventType string      `json:"event_type"`
	OccuredAt time.Time   `json:"occurred_at"`
	Payload   interface{} `json:"payload"`
}

func NewDomainEvent(eventType string, payload interface{}) DomainEvent {
	return DomainEvent{
		EventID:   uuid.New(),
		EventType: eventType,
		OccuredAt: time.Now(),
		Payload:   payload,
	}
}

type AIScore struct {
	Score      float64            `json:"score"`       // 0–100
	Confidence float64            `json:"confidence"`  // 0–1
	Breakdown  map[string]float64 `json:"breakdown"`
	ScoredAt   time.Time          `json:"scored_at"`
	ModelID    string             `json:"model_id"`
}

type Money struct {
	Amount   int64  `json:"amount"` // stored in minor units (VND, cents…)
	Currency string `json:"currency"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
}
