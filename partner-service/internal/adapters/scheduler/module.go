package scheduler

import (
	"context"

	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC  fx.Lifecycle
	Svc *services.PartnerService
}

func New(p Params) {
	s := NewPartnerScheduler(*p.Svc)

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return s.Start()
		},
		OnStop: func(context.Context) error {
			s.Stop()
			return nil
		},
	})
}

var Module = fx.Options(
	fx.Invoke(New),
)
