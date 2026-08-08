package telegram

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/bot-service/config"
	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type contentPublisher struct {
	bot *tgbotapi.BotAPI
	cfg config.TelegramConfig
}

func NewContentPublisher(bot *tgbotapi.BotAPI, cfg *config.Config) port.ChannelPublisher {
	return &contentPublisher{bot: bot, cfg: cfg.Telegram}
}

func (p *contentPublisher) Channel() model.Channel { return model.ChannelTelegram }

func (p *contentPublisher) Publish(ctx context.Context, post model.Post) (string, error) {
	if p.cfg.BroadcastChatID == "" {
		return "", fmt.Errorf("telegram: BroadcastChatID not configured")
	}
	chatID, err := strconv.ParseInt(p.cfg.BroadcastChatID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("telegram: invalid BroadcastChatID %q: %w", p.cfg.BroadcastChatID, err)
	}

	if len(post.Media) > 0 {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(post.Media[0].URL))
		photo.Caption = post.Text
		sent, err := p.bot.Send(photo)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(sent.MessageID), nil
	}

	sent, err := p.bot.Send(tgbotapi.NewMessage(chatID, post.Text))
	if err != nil {
		return "", err
	}
	return strconv.Itoa(sent.MessageID), nil
}
