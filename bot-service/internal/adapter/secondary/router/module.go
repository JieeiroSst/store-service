package router

import (
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	Senders []port.ChannelSender `group:"channel_senders"`
}

func NewRouter(p Params) port.MessageSender {
	return New(p.Senders)
}

var Module = fx.Options(
	fx.Provide(NewRouter),
)
