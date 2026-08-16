package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/calculate-service/config"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

type MarketParams struct {
	fx.In

	LC     fx.Lifecycle
	Cfg    *config.Config
	Market port.MarketService
}

func NewMarket(p MarketParams) {
	interval := p.Cfg.Market.RefreshIntervalSec
	if interval <= 0 {
		interval = 15
	}

	c := cron.New()
	_, _ = c.AddFunc(fmt.Sprintf("@every %ds", interval), func() {
		refreshMarket(p.Market)
	})

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			c.Start()
			go refreshMarket(p.Market)
			return nil
		},
		OnStop: func(context.Context) error {
			<-c.Stop().Done()
			return nil
		},
	})
}

func refreshMarket(market port.MarketService) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := market.RefreshMarkets(ctx); err != nil {
		logger.Error(ctx, "RefreshMarkets error %v", err)
	}
}
