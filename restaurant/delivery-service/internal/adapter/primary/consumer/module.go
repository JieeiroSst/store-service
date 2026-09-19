package consumer

import (
	"context"

	"go.uber.org/fx"
)

func registerLifecycle(lc fx.Lifecycle, s *Subscriber) {
	lc.Append(fx.Hook{
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
	fx.Provide(NewSubscriber),
	fx.Invoke(registerLifecycle),
)
