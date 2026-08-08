package twitter

import (
	"context"
	"time"

	"github.com/JIeeiroSst/bot-service/config"
	secondary "github.com/JIeeiroSst/bot-service/internal/adapter/secondary/twitter"
	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/zap"
)

type Poller struct {
	client  *secondary.Client
	usecase port.BotUsecase
	cfg     config.TwitterConfig
	sinceID string
}

func NewPoller(client *secondary.Client, usecase port.BotUsecase, cfg *config.Config) *Poller {
	return &Poller{client: client, usecase: usecase, cfg: cfg.Twitter}
}

func (p *Poller) Run(ctx context.Context) {
	if !p.cfg.Enabled {
		return
	}
	lg := logger.WithContext(ctx)

	interval := time.Duration(p.cfg.PollIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mentions, newSinceID, err := p.client.FetchMentions(ctx, p.cfg.UserID, p.sinceID)
			if err != nil {
				lg.Error("twitter poller: fetch mentions", zap.Error(err))
				continue
			}
			p.sinceID = newSinceID

			for _, m := range mentions {
				msg := model.IncomingMessage{
					Channel:      model.ChannelTwitter,
					ChatID:       m.TweetID,
					MessageID:    m.TweetID,
					FromUsername: m.AuthorID,
					Text:         m.Text,
				}
				if err := p.usecase.HandleMessage(ctx, msg); err != nil {
					lg.Error("twitter poller: handle message", zap.Error(err))
				}
			}
		}
	}
}
