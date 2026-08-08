package application

import (
	"context"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/zap"
)

type botService struct {
	sender port.MessageSender
}

func NewBotService(sender port.MessageSender) port.BotUsecase {
	return &botService{sender: sender}
}

func (s *botService) HandleMessage(ctx context.Context, msg model.IncomingMessage) error {
	lg := logger.WithContext(ctx)

	if msg.Text == "" {
		return nil
	}

	lg.Info("received message",
		zap.String("channel", string(msg.Channel)),
		zap.String("chat_id", msg.ChatID),
		zap.String("from", msg.FromUsername),
	)

	err := s.sender.SendMessage(ctx, model.OutgoingMessage{
		Channel:          msg.Channel,
		ChatID:           msg.ChatID,
		Text:             msg.Text,
		ReplyToMessageID: msg.MessageID,
	})
	if err != nil {
		lg.Error("HandleMessage: send reply", zap.Error(err))
		return err
	}
	return nil
}
