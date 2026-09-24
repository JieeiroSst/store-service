package main

import (
	"testing"

	"go.uber.org/fx"

	"github.com/JIeeiroSst/bonuslink-service/internal/adapters/primary/http"
	"github.com/JIeeiroSst/bonuslink-service/internal/adapters/primary/queue"
	"github.com/JIeeiroSst/bonuslink-service/internal/adapters/secondary/postgres"
	"github.com/JIeeiroSst/bonuslink-service/internal/config"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/services"
	"github.com/JIeeiroSst/bonuslink-service/pkg/logger"
)

// TestDependencyGraph fails if any constructor's dependencies are unsatisfied.
func TestDependencyGraph(t *testing.T) {
	err := fx.ValidateApp(
		config.Module,
		fx.Provide(newLoggerConfig),
		logger.Module,
		postgres.Module,
		services.Module,
		http.Module,
		http.ServerModule,
		queue.Module,
	)
	if err != nil {
		t.Fatal(err)
	}
}
