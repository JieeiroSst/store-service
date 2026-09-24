package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type stubExpiry struct{ calls atomic.Int32 }

func (s *stubExpiry) ExpireOverdue(context.Context) (int64, error) { s.calls.Add(1); return 0, nil }

func params(spec string, uc *stubExpiry, lc fx.Lifecycle) Params {
	return Params{LC: lc, Config: &config.Config{Scheduler: config.SchedulerConfig{ContractExpiryCron: spec}}, Contract: uc}
}

func TestRegister_RejectsBadSpec(t *testing.T) {
	if err := Register(params("not a cron", &stubExpiry{}, fxtest.NewLifecycle(t))); err == nil {
		t.Fatal("expected an error for an invalid cron spec")
	}
}

func TestRegister_DisabledRegistersNoHooks(t *testing.T) {
	for _, spec := range []string{"", "off", "OFF"} {
		lc := fxtest.NewLifecycle(t)
		if err := Register(params(spec, &stubExpiry{}, lc)); err != nil {
			t.Fatal(err)
		}
		lc.RequireStart().RequireStop() // would hang or run the job if hooks existed
	}
}

func TestRegister_RunsOnceOnStart(t *testing.T) {
	uc := &stubExpiry{}
	lc := fxtest.NewLifecycle(t)
	if err := Register(params("0 0 1 1 *", uc, lc)); err != nil {
		t.Fatal(err)
	}
	lc.RequireStart().RequireStop()
	// The catch-up run is async; RequireStop waits for the cron runner, not it.
	for i := 0; i < 100 && uc.calls.Load() == 0; i++ {
		waitBriefly()
	}
	if uc.calls.Load() != 1 {
		t.Fatalf("catch-up run count = %d, want 1", uc.calls.Load())
	}
}

func waitBriefly() { time.Sleep(10 * time.Millisecond) }
