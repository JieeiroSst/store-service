package application

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
	"github.com/JIeeiroSst/serpapi-service/internal/port"
)

type AccountService struct {
	api          port.SerpAPI
	cache        port.Cache
	metrics      port.Metrics
	limiter      *Limiter
	locationsTTL time.Duration
}

func NewAccountService(cfg *config.Config, api port.SerpAPI, cache port.Cache, metrics port.Metrics, limiter *Limiter) *AccountService {
	return &AccountService{api: api, cache: cache, metrics: metrics, limiter: limiter, locationsTTL: cfg.Cache.LocationsTTL}
}

func (s *AccountService) Account(ctx context.Context) (domain.Account, error) {
	release, err := s.limiter.Acquire(ctx)
	if err != nil {
		return domain.Account{}, err
	}
	defer release()
	start := time.Now()
	acc, err := s.api.Account(ctx)
	s.metrics.ObserveUpstream("account", "", outcome(err), time.Since(start))
	return acc, err
}

func (s *AccountService) Locations(ctx context.Context, q domain.LocationQuery) ([]domain.Location, error) {
	q.Q = strings.TrimSpace(q.Q)
	if err := q.Normalize(); err != nil {
		return nil, err
	}
	key := "locations:" + strconv.Itoa(q.Limit) + ":" + strings.ToLower(q.Q)
	if raw, ok, err := s.cache.Get(ctx, key); err == nil && ok {
		var locs []domain.Location
		if json.Unmarshal(raw, &locs) == nil {
			s.metrics.ObserveCache("locations", true)
			return locs, nil
		}
	} else if err != nil {
		log.Printf("cache get locations: %v", err)
	}
	s.metrics.ObserveCache("locations", false)

	release, err := s.limiter.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	start := time.Now()
	locs, err := s.api.Locations(ctx, q)
	s.metrics.ObserveUpstream("locations", "", outcome(err), time.Since(start))
	if err != nil {
		return nil, err
	}
	if locs == nil {
		locs = []domain.Location{}
	}
	if raw, err := json.Marshal(locs); err == nil {
		if err := s.cache.Set(ctx, key, raw, s.locationsTTL); err != nil {
			log.Printf("cache set locations: %v", err)
		}
	}
	return locs, nil
}
