package consul

import (
	"context"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	consulapi "github.com/hashicorp/consul/api"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func RegisterService(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) {
	if !cfg.ConsulEnabled || cfg.HostConsul == "" {
		log.Info("consul registration disabled")
		return
	}

	clientCfg := consulapi.DefaultConfig()
	clientCfg.Address = strings.TrimPrefix(strings.TrimPrefix(cfg.HostConsul, "http://"), "https://")
	client, err := consulapi.NewClient(clientCfg)
	if err != nil {
		log.Warn("consul client init failed, continuing without registration", zap.Error(err))
		return
	}

	port, _ := strconv.Atoi(cfg.Port)
	serviceID := cfg.KeyConsul + "-" + cfg.Port

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := client.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
				ID:   serviceID,
				Name: cfg.KeyConsul,
				Port: port,
			})
			if err != nil {
				log.Warn("consul service registration failed, continuing without it", zap.Error(err))
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			_ = client.Agent().ServiceDeregister(serviceID)
			return nil
		},
	})
}
