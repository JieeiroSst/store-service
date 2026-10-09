package port

import (
	"context"

	"github.com/JIeeiroSst/search-service/internal/domain"
)

type DocumentSearcher interface {
	Query(ctx context.Context, q domain.DocumentQuery, mode domain.QueryMode) (domain.DocumentPage, error)
	Autocomplete(ctx context.Context, q domain.AutocompleteQuery) ([]domain.Document, error)
	Get(ctx context.Context, ref domain.DocumentRef) (domain.Document, error)
	Similar(ctx context.Context, ref domain.DocumentRef, size int) ([]domain.Document, error)
	Indices(ctx context.Context) ([]domain.IndexInfo, error)
	Fields(ctx context.Context, index string) ([]domain.FieldInfo, error)
}

type SchemaManager interface {
	Apply(ctx context.Context) error
	Status(ctx context.Context) (domain.SchemaStatus, error)
	Reindex(ctx context.Context, name string) (domain.ReindexResult, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
