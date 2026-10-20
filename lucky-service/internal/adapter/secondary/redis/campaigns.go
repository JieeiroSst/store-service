package rediscache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
)

type Campaigns struct {
	port.CampaignRepository
	rdb       *redis.Client
	key       string
	ttl       time.Duration
	opTimeout time.Duration
}

func NewCampaigns(base port.CampaignRepository, rdb *redis.Client, prefix string, ttl, opTimeout time.Duration) *Campaigns {
	return &Campaigns{CampaignRepository: base, rdb: rdb, key: prefix + "campaigns:running", ttl: ttl, opTimeout: opTimeout}
}

func (c *Campaigns) op(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), c.opTimeout)
}

func (c *Campaigns) Running(ctx context.Context, at time.Time) ([]model.Campaign, error) {
	rctx, cancel := c.op(ctx)
	raw, err := c.rdb.Get(rctx, c.key).Bytes()
	cancel()
	if err == nil {
		var cached []model.Campaign
		if json.Unmarshal(raw, &cached) == nil {
			return cached, nil
		}
	} else if err != redis.Nil {
		slog.Warn("campaign cache read failed", "err", err)
	}
	fresh, err := c.CampaignRepository.Running(ctx, at)
	if err != nil {
		return nil, err
	}
	if raw, err := json.Marshal(fresh); err == nil {
		wctx, cancel := c.op(ctx)
		if err := c.rdb.Set(wctx, c.key, raw, c.ttl).Err(); err != nil {
			slog.Warn("campaign cache write failed", "err", err)
		}
		cancel()
	}
	return fresh, nil
}

func (c *Campaigns) invalidate(ctx context.Context) {
	rctx, cancel := c.op(ctx)
	defer cancel()
	if err := c.rdb.Del(rctx, c.key).Err(); err != nil {
		slog.Warn("campaign cache invalidation failed, entries expire on their own", "err", err)
	}
}

func (c *Campaigns) Update(ctx context.Context, m *model.Campaign) error {
	err := c.CampaignRepository.Update(ctx, m)
	c.invalidate(ctx)
	return err
}

func (c *Campaigns) Delete(ctx context.Context, id int64) error {
	err := c.CampaignRepository.Delete(ctx, id)
	c.invalidate(ctx)
	return err
}

func (c *Campaigns) SetStatus(ctx context.Context, id int64, status model.CampaignStatus) error {
	err := c.CampaignRepository.SetStatus(ctx, id, status)
	c.invalidate(ctx)
	return err
}
