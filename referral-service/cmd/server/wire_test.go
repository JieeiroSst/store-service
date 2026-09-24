package main

import (
	"testing"

	"go.uber.org/fx"

	"github.com/referral/service/internal/adapters/primary/http"
	"github.com/referral/service/internal/adapters/secondary/cache"
	"github.com/referral/service/internal/adapters/secondary/mysql"
	"github.com/referral/service/internal/adapters/secondary/queue"
	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/services"
	"github.com/referral/service/pkg/logger"
)

// TestDependencyGraph fails if any constructor's dependencies are unsatisfied.
func TestDependencyGraph(t *testing.T) {
	if err := fx.ValidateApp(
		config.Module,
		fx.Provide(newLoggerConfig),
		logger.Module,
		mysql.Module,
		cache.Module,
		queue.Module,
		services.Module,
		services.RelayModule,
		http.Module,
		http.ServerModule,
	); err != nil {
		t.Fatal(err)
	}
}
