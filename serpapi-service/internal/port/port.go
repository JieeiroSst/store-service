package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

type SerpAPI interface {
	Search(ctx context.Context, req domain.SearchRequest) (domain.SearchResult, error)
	Archive(ctx context.Context, searchID string, output domain.Output) (domain.SearchResult, error)
	Account(ctx context.Context) (domain.Account, error)
	Locations(ctx context.Context, q domain.LocationQuery) ([]domain.Location, error)
	UploadImage(ctx context.Context, img domain.ImageUpload) (domain.UploadedImage, error)
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type Metrics interface {
	ObserveUpstream(op, engine, outcome string, d time.Duration)
	ObserveCache(op string, hit bool)
}
