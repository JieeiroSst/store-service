package app

import (
	"context"

	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

type HealthService struct {
	db    ports.DBPing
	cache ports.Cache
}

func NewHealthService(db ports.DBPing, cache ports.Cache) *HealthService {
	return &HealthService{db, cache}
}

type ReadyResult struct {
	Ready    bool
	Database string // "ok" | "error"
	Redis    string // "ok" | "error" | "" (not configured)
}

func (s *HealthService) Ready(ctx context.Context) ReadyResult {
	result := ReadyResult{Database: "ok"}
	if err := s.db.Ping(ctx); err != nil {
		result.Database = "error"
	}
	if s.cache.Enabled() {
		if err := s.cache.Ping(ctx); err != nil {
			result.Redis = "error"
		} else {
			result.Redis = "ok"
		}
	}
	result.Ready = result.Database == "ok"
	return result
}
