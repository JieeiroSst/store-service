package facebook

import (
	"context"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
)

type sender struct {
	client *Client
}

func NewMessageSender(client *Client) port.ChannelSender {
	return &sender{client: client}
}

func (s *sender) Channel() model.Channel { return model.ChannelFacebook }

func (s *sender) SendMessage(ctx context.Context, msg model.OutgoingMessage) error {
	return s.client.SendText(ctx, msg.ChatID, msg.Text)
}
