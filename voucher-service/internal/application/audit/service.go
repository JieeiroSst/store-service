package audit

import (
	"context"
	"time"
)

type Service struct {
	repo AuditRepository
}

func NewService(repo AuditRepository) AuditService {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, entry Entry) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	return s.repo.Insert(ctx, entry)
}

func (s *Service) Query(ctx context.Context, in QueryInput) ([]Entry, error) {
	return s.repo.Query(ctx, in)
}
