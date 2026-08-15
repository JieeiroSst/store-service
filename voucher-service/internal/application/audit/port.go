package audit

import (
	"context"
	"time"
)

type Entry struct {
	ActorType  string
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Before     map[string]any
	After      map[string]any
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
}

type QueryInput struct {
	EntityType string
	EntityID   string
	Limit      int
}

type AuditService interface {
	Record(ctx context.Context, entry Entry) error
	Query(ctx context.Context, in QueryInput) ([]Entry, error)
}

type AuditRepository interface {
	Insert(ctx context.Context, entry Entry) error
	Query(ctx context.Context, in QueryInput) ([]Entry, error)
}
