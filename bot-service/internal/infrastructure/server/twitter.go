package server

import (
	"context"

	twitterprimary "github.com/JIeeiroSst/bot-service/internal/adapter/primary/twitter"
	"go.uber.org/fx"
)

type TwitterParams struct {
	fx.In

	LC     fx.Lifecycle
	Poller *twitterprimary.Poller
}

func NewTwitterServer(p TwitterParams) {
	ctx, cancel := context.WithCancel(context.Background())

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go p.Poller.Run(ctx)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}
