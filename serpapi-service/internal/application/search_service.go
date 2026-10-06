package application

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
	"github.com/JIeeiroSst/serpapi-service/internal/port"
)

type SearchService struct {
	catalog  *domain.Catalog
	api      port.SerpAPI
	cache    port.Cache
	metrics  port.Metrics
	limiter  *Limiter
	strict   bool
	ttl      time.Duration
	maxBatch int
}

func NewSearchService(
	cfg *config.Config,
	catalog *domain.Catalog,
	api port.SerpAPI,
	cache port.Cache,
	metrics port.Metrics,
	limiter *Limiter,
) *SearchService {
	return &SearchService{
		catalog:  catalog,
		api:      api,
		cache:    cache,
		metrics:  metrics,
		limiter:  limiter,
		strict:   cfg.Search.StrictEngines,
		ttl:      cfg.Cache.TTL,
		maxBatch: cfg.Search.MaxBatchSize,
	}
}

func (s *SearchService) Engines(group string) []domain.Engine { return s.catalog.List(group) }

func (s *SearchService) Engine(id string) (domain.Engine, error) {
	e, ok := s.catalog.Get(id)
	if !ok {
		return e, domain.ErrNotFound
	}
	return e, nil
}

func (s *SearchService) Search(ctx context.Context, req domain.SearchRequest) (domain.SearchResult, error) {
	req, err := s.prepare(req)
	if err != nil {
		return domain.SearchResult{}, err
	}

	cacheable := !req.Async && !req.ZeroTrace()
	key := req.CacheKey()
	if cacheable && !req.NoCache {
		if res, ok := s.cached(ctx, "search", key); ok {
			return res, nil
		}
	}

	var res domain.SearchResult
	err = s.call(ctx, "search", s.engineLabel(req.Engine), func(ctx context.Context) error {
		var err error
		res, err = s.api.Search(ctx, req)
		return err
	})
	if err != nil {
		return res, err
	}
	if cacheable && (!req.Output.IsJSON() || res.Status == domain.StatusSuccess) {
		s.store(ctx, key, res, s.ttl)
	}
	return res, nil
}

type BatchItem struct {
	Result domain.SearchResult
	Err    error
}

func (s *SearchService) Batch(ctx context.Context, reqs []domain.SearchRequest) ([]BatchItem, error) {
	if len(reqs) == 0 {
		return nil, domain.Invalid("batch needs at least one search")
	}
	if len(reqs) > s.maxBatch {
		return nil, domain.Invalid("batch holds at most %d searches, got %d", s.maxBatch, len(reqs))
	}
	out := make([]BatchItem, len(reqs))
	var wg sync.WaitGroup
	for i, r := range reqs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := s.Search(ctx, r)
			out[i] = BatchItem{Result: res, Err: err}
		}()
	}
	wg.Wait()
	return out, nil
}

func (s *SearchService) Archive(ctx context.Context, id string, output domain.Output) (domain.SearchResult, error) {
	if err := domain.ValidateSearchID(id); err != nil {
		return domain.SearchResult{}, err
	}
	if output == domain.OutputMarkdown {
		return domain.SearchResult{}, domain.Invalid("the search archive serves json, html or json_with_pixel_position")
	}
	key := "archive:" + string(output) + ":" + id
	if res, ok := s.cached(ctx, "archive", key); ok {
		return res, nil
	}
	var res domain.SearchResult
	err := s.call(ctx, "archive", "", func(ctx context.Context) error {
		var err error
		res, err = s.api.Archive(ctx, id, output)
		return err
	})
	if err != nil {
		return res, err
	}

	if res.Status == domain.StatusSuccess {
		s.store(ctx, key, res, s.ttl)
	}
	return res, nil
}

func (s *SearchService) prepare(req domain.SearchRequest) (domain.SearchRequest, error) {
	req.Engine = strings.TrimSpace(req.Engine)
	if req.Engine == "" {
		return req, domain.Invalid("engine is required")
	}
	if !domain.ValidEngineID(req.Engine) {
		return req, domain.Invalid("invalid engine %q", req.Engine)
	}
	if req.Output == "" {
		req.Output = domain.OutputJSON
	}
	if req.Async && req.NoCache {
		return req, domain.Invalid("async cannot be combined with no_cache")
	}
	if req.Output == domain.OutputPixelPosition && !domain.SupportsPixelPosition(req.Engine) {
		return req, domain.Invalid("output json_with_pixel_position is only available for google and google_ads")
	}
	params := make(domain.Params, len(req.Params))
	for k, v := range req.Params {
		if k == "" || domain.IsReservedParam(k) {
			continue
		}
		params[k] = v
	}
	req.Params = params

	e, ok := s.catalog.Get(req.Engine)
	if !ok {
		if s.strict {
			return req, domain.Invalid("unknown engine %q, see GET /api/v1/engines", req.Engine)
		}
		return req, nil
	}
	return req, e.Validate(params)
}

func (s *SearchService) engineLabel(id string) string {
	if _, ok := s.catalog.Get(id); ok {
		return id
	}
	return "other"
}

func (s *SearchService) call(ctx context.Context, op, engine string, fn func(context.Context) error) error {
	release, err := s.limiter.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	start := time.Now()
	err = fn(ctx)
	s.metrics.ObserveUpstream(op, engine, outcome(err), time.Since(start))
	return err
}

func (s *SearchService) cached(ctx context.Context, op, key string) (domain.SearchResult, bool) {
	raw, ok, err := s.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get %s: %v", op, err)
		return domain.SearchResult{}, false
	}
	if ok {
		res, err := decodeResult(raw)
		if err == nil {
			s.metrics.ObserveCache(op, true)
			res.Cached = true
			return res, true
		}
		log.Printf("cache decode %s: %v", op, err)
	}
	s.metrics.ObserveCache(op, false)
	return domain.SearchResult{}, false
}

func (s *SearchService) store(ctx context.Context, key string, res domain.SearchResult, ttl time.Duration) {
	if err := s.cache.Set(ctx, key, encodeResult(res), ttl); err != nil {
		log.Printf("cache set: %v", err)
	}
}

func outcome(err error) string {
	switch {
	case err == nil:
		return "ok"
	case domain.IsInvalid(err):
		return "invalid"
	case errors.Is(err, domain.ErrNotFound):
		return "not_found"
	case errors.Is(err, domain.ErrRateLimited):
		return "rate_limited"
	case errors.Is(err, domain.ErrExpired):
		return "expired"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return "canceled"
	default:
		return "upstream"
	}
}
