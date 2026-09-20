package cron

import (
	"context"
	"fmt"
	"log"

	"github.com/JIeeiroSst/coupon-service/config"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
	"github.com/robfig/cron"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Coupons port.CouponUsecase
	Config  *config.Config
}

func New(p Params) error {
	scheduler := cron.New()

	err := scheduler.AddFunc(p.Config.Cron.DeactivateStaleSchedule, func() {
		n, err := p.Coupons.DeactivateStaleCoupons(context.Background())
		if err != nil {
			log.Printf("deactivate stale coupons: %v", err)
			return
		}
		if n > 0 {
			log.Printf("deactivated %d stale coupon(s)", n)
		}
	})
	if err != nil {
		return fmt.Errorf("schedule deactivate-stale-coupons job %q: %w", p.Config.Cron.DeactivateStaleSchedule, err)
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			scheduler.Start()
			return nil
		},
		OnStop: func(context.Context) error {
			scheduler.Stop()
			return nil
		},
	})

	return nil
}
