package telegram

import (
	"github.com/JIeeiroSst/bot-service/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewBotAPI(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, err
	}
	bot.Debug = cfg.Telegram.Debug
	return bot, nil
}
