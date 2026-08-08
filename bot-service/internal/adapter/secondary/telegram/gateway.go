package telegram

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type sender struct {
	bot *tgbotapi.BotAPI
}

func NewMessageSender(bot *tgbotapi.BotAPI) port.ChannelSender {
	return &sender{bot: bot}
}

func (s *sender) Channel() model.Channel { return model.ChannelTelegram }

func (s *sender) SendMessage(ctx context.Context, msg model.OutgoingMessage) error {
	chatID, err := strconv.ParseInt(msg.ChatID, 10, 64)
	if err != nil {
		return fmt.Errorf("telegram: invalid chat id %q: %w", msg.ChatID, err)
	}

	out := tgbotapi.NewMessage(chatID, msg.Text)
	if msg.ReplyToMessageID != "" {
		if replyTo, err := strconv.Atoi(msg.ReplyToMessageID); err == nil {
			out.ReplyToMessageID = replyTo
		}
	}
	_, err = s.bot.Send(out)
	return err
}
