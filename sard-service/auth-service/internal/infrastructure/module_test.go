package infrastructure

import (
	"context"
	"testing"
	"time"

	"go.uber.org/fx"
)

func TestModuleGraph(t *testing.T) {
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatal(err)
	}
}

func TestAppStartsWithMemoryBackend(t *testing.T) {
	t.Setenv("CONSUL_ADDR", "")
	t.Setenv("STORAGE_BACKEND", "memory")
	t.Setenv("PORT_HTTP_SERVER", "0")
	t.Setenv("STORAGE_BACKEND", "memory")
	app := fx.New(Module, fx.NopLogger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := app.Stop(ctx); err != nil {
		t.Fatal(err)
	}
}
