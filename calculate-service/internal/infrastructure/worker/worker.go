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

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Cfg     *config.Config
	Weather port.WeatherService
}

func New(p Params) {
	interval := p.Cfg.Weather.RefreshIntervalMin
	if interval <= 0 {
		interval = 15
	}

	c := cron.New()
	_, _ = c.AddFunc(fmt.Sprintf("@every %dm", interval), func() {
		refresh(p.Weather)
	})

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			c.Start()
			go refresh(p.Weather)
			return nil
		},
		OnStop: func(context.Context) error {
			<-c.Stop().Done()
			return nil
		},
	})
}

func refresh(weather port.WeatherService) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := weather.RefreshTrackedLocations(ctx); err != nil {
		logger.Error(ctx, "RefreshTrackedLocations error %v", err)
	}
}
