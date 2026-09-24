package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

const jobTimeout = 2 * time.Minute

type Params struct {
	fx.In

	LC       fx.Lifecycle
	Config   *config.Config
	Contract port.ContractExpiryUsecase
}

func Register(p Params) error {
	spec := strings.TrimSpace(p.Config.Scheduler.ContractExpiryCron)
	if spec == "" || strings.EqualFold(spec, "off") {
		return nil
	}

	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger), cron.SkipIfStillRunning(cron.DiscardLogger)))
	if _, err := c.AddFunc(spec, func() { expireContracts(p.Contract) }); err != nil {
		return fmt.Errorf("invalid CONTRACT_EXPIRY_CRON %q: %w", spec, err)
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			c.Start()
			go expireContracts(p.Contract)
			logrus.Infof("contract expiry job scheduled (%s)", spec)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			select {
			case <-c.Stop().Done():
			case <-ctx.Done():
			}
			return nil
		},
	})
	return nil
}

func expireContracts(uc port.ContractExpiryUsecase) {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	n, err := uc.ExpireOverdue(ctx)
	if err != nil {
		logrus.WithError(err).Error("contract expiry job failed")
		return
	}
	logrus.WithField("expired", n).Info("contract expiry job finished")
}
