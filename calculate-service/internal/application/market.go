package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/calculate-service/common"
	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
)

const marketSnapshotCacheKey = "market:snapshot"

const marketSnapshotTTL = 2 * time.Minute

type MarketServiceConfig struct {
	VsCurrency string
	PerPage    int
}

type marketService struct {
	provider    port.MarketProvider
	cache       port.MarketCache
	broadcaster port.MarketBroadcaster
	cfg         MarketServiceConfig
}

func NewMarketService(
	provider port.MarketProvider,
	cache port.MarketCache,
	broadcaster port.MarketBroadcaster,
	cfg MarketServiceConfig,
) port.MarketService {
	return &marketService{
		provider:    provider,
		cache:       cache,
		broadcaster: broadcaster,
		cfg:         cfg,
	}
}

func (s *marketService) GetSnapshot(ctx context.Context) (*model.MarketSnapshot, error) {
	if cached, ok := s.cache.GetMarketSnapshot(ctx, marketSnapshotCacheKey); ok {
		return cached, nil
	}
	return nil, common.ErrMarketUnavailable
}

func (s *marketService) RefreshMarkets(ctx context.Context) error {
	coins, err := s.provider.GetMarkets(ctx, s.cfg.VsCurrency, s.cfg.PerPage)
	if err != nil {
		logger.Error(ctx, "RefreshMarkets upstream error %v", err)
		return fmt.Errorf("%w: %v", common.ErrMarketUnavailable, err)
	}

	snapshot := &model.MarketSnapshot{
		VsCurrency:  s.cfg.VsCurrency,
		GeneratedAt: time.Now().UTC(),
		Coins:       coins,
	}

	s.cache.SetMarketSnapshot(ctx, marketSnapshotCacheKey, snapshot, marketSnapshotTTL)
	s.broadcaster.Broadcast(snapshot)
	return nil
}
