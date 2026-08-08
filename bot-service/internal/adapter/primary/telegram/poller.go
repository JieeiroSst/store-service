package telegram

import (
	"context"
	"strconv"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Poller struct {
	bot     *tgbotapi.BotAPI
	usecase port.BotUsecase
}

func NewPoller(bot *tgbotapi.BotAPI, usecase port.BotUsecase) *Poller {
	return &Poller{bot: bot, usecase: usecase}
}

func (p *Poller) Run(ctx context.Context) {
	lg := logger.WithContext(ctx)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := p.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			p.bot.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Message == nil {
				continue
			}

			msg := model.IncomingMessage{
				Channel:   model.ChannelTelegram,
				ChatID:    strconv.FormatInt(update.Message.Chat.ID, 10),
				MessageID: strconv.Itoa(update.Message.MessageID),
				Text:      update.Message.Text,
			}
			if update.Message.From != nil {
				msg.FromUsername = update.Message.From.UserName
			}

			if err := p.usecase.HandleMessage(ctx, msg); err != nil {
				lg.Error("poller: handle message", zap.Error(err))
			}
		}
	}
}
