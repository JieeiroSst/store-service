package server

import (
	"context"

	telegramadapter "github.com/JIeeiroSst/bot-service/internal/adapter/primary/telegram"
	"go.uber.org/fx"
)

type TelegramParams struct {
	fx.In

	LC     fx.Lifecycle
	Poller *telegramadapter.Poller
}

func NewTelegramServer(p TelegramParams) {
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
