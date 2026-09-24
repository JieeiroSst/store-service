package infrastructure

import (
	"testing"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// TestDependencyGraph fails if any constructor's dependencies are unsatisfied.
func TestDependencyGraph(t *testing.T) {
	err := fx.ValidateApp(
		fx.Provide(func() *config.Config { return &config.Config{} }),
		fx.Provide(func() *gorm.DB { return nil }),
		Core,
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestDependencyGraph_WithTemporal builds the graph with Temporal switched on:
// the lazy client, the orchestrator and the worker registration. Nothing is
// started, so no server is needed.
func TestDependencyGraph_WithTemporal(t *testing.T) {
	cfg := &config.Config{Temporal: config.TemporalConfig{Address: "localhost:1", Namespace: "default", TaskQueue: "q"}}
	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() *config.Config { return cfg }),
		fx.Provide(func() *gorm.DB { return nil }),
		Core,
	)
	if err := app.Err(); err != nil {
		t.Fatal(err)
	}
}
