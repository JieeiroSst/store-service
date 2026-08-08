package router

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
)

type Router struct {
	senders map[model.Channel]port.ChannelSender
}

func New(senders []port.ChannelSender) *Router {
	m := make(map[model.Channel]port.ChannelSender, len(senders))
	for _, s := range senders {
		m[s.Channel()] = s
	}
	return &Router{senders: m}
}

func (r *Router) SendMessage(ctx context.Context, msg model.OutgoingMessage) error {
	sender, ok := r.senders[msg.Channel]
	if !ok {
		return fmt.Errorf("router: no sender registered for channel %q", msg.Channel)
	}
	return sender.SendMessage(ctx, msg)
}
