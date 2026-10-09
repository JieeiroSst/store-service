package application

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/JIeeiroSst/search-service/internal/domain"
	"github.com/JIeeiroSst/search-service/internal/port"
)

const (
	applyRetryMin = 2 * time.Second
	applyRetryMax = time.Minute
)

type SchemaService struct {
	schema port.SchemaManager
}

func NewSchemaService(schema port.SchemaManager) *SchemaService {
	return &SchemaService{schema: schema}
}

func (s *SchemaService) Apply(ctx context.Context) (domain.SchemaStatus, error) {
	if err := s.schema.Apply(ctx); err != nil {
		return domain.SchemaStatus{}, err
	}
	return s.schema.Status(ctx)
}

func (s *SchemaService) Status(ctx context.Context) (domain.SchemaStatus, error) {
	return s.schema.Status(ctx)
}

func (s *SchemaService) Reindex(ctx context.Context, names []string, unmanaged bool) ([]domain.ReindexResult, error) {
	names = domain.SplitList(names)
	for _, name := range names {
		if err := domain.ValidateIndex(name, false); err != nil {
			return nil, err
		}
	}
	if unmanaged {
		status, err := s.schema.Status(ctx)
		if err != nil {
			return nil, err
		}
		for _, idx := range status.Indices {
			if !idx.Managed || idx.WriteBlocked {
				names = append(names, idx.Name)
			}
		}
	}
	if len(names) == 0 {
		return nil, domain.Invalid("index or unmanaged=true is required")
	}

	out := make([]domain.ReindexResult, 0, len(names))
	for _, name := range names {
		res, err := s.schema.Reindex(ctx, name)
		if err != nil {
			res.Name = name
			res.Error = err.Error()
		}
		out = append(out, res)
	}
	return out, nil
}

func (s *SchemaService) ApplyWithRetry(ctx context.Context) {
	backoff := applyRetryMin
	for {
		err := s.schema.Apply(ctx)
		if err == nil {
			log.Printf("elasticsearch schema applied")
			s.warnBlocked(ctx)
			return
		}
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return
		}
		log.Printf("apply elasticsearch schema failed, retrying in %s: %v", backoff, err)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}
		if backoff *= 2; backoff > applyRetryMax {
			backoff = applyRetryMax
		}
	}
}

func (s *SchemaService) warnBlocked(ctx context.Context) {
	status, err := s.schema.Status(ctx)
	if err != nil {
		return
	}
	for _, idx := range status.Indices {
		if idx.WriteBlocked {
			log.Printf("index %s is write-blocked, likely from an interrupted reindex: run POST /api/v1/admin/reindex?index=%s", idx.Index, idx.Name)
		}
	}
}
