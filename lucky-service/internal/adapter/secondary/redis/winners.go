package rediscache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/redis/go-redis/v9"
)

const winnersRetention = 30 * 24 * time.Hour

type Winners struct {
	rdb       *redis.Client
	prefix    string
	size      int
	opTimeout time.Duration
}

func NewWinners(rdb *redis.Client, prefix string, size int, opTimeout time.Duration) *Winners {
	return &Winners{rdb: rdb, prefix: prefix, size: max(size, 1), opTimeout: opTimeout}
}

func (w *Winners) key(campaignID int64) string {
	return w.prefix + "winners:" + strconv.FormatInt(campaignID, 10)
}

func (w *Winners) Push(ctx context.Context, win model.Winner) error {
	raw, err := json.Marshal(win)
	if err != nil {
		return err
	}
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), w.opTimeout)
	defer cancel()
	key := w.key(win.CampaignID)
	_, err = w.rdb.TxPipelined(c, func(p redis.Pipeliner) error {
		p.LPush(c, key, raw)
		p.LTrim(c, key, 0, int64(w.size-1))
		p.Expire(c, key, winnersRetention)
		return nil
	})
	return err
}

func (w *Winners) Recent(ctx context.Context, campaignID int64, limit int) ([]model.Winner, error) {
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), w.opTimeout)
	defer cancel()
	items, err := w.rdb.LRange(c, w.key(campaignID), 0, int64(min(limit, w.size)-1)).Result()
	if err != nil {
		return nil, err
	}
	out := make([]model.Winner, 0, len(items))
	for _, it := range items {
		var win model.Winner
		if json.Unmarshal([]byte(it), &win) == nil {
			out = append(out, win)
		}
	}
	return out, nil
}
