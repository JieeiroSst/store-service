package rediscache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/redis/go-redis/v9"
)

var signalWindows = map[model.SignalKind]time.Duration{
	model.SignalDevice: 30 * 24 * time.Hour,
	model.SignalIP:     24 * time.Hour,
}

type Signals struct {
	rdb       *redis.Client
	prefix    string
	opTimeout time.Duration
}

func NewSignals(rdb *redis.Client, prefix string, opTimeout time.Duration) *Signals {
	return &Signals{rdb: rdb, prefix: prefix, opTimeout: opTimeout}
}

func (s *Signals) key(kind model.SignalKind, value string) string {
	sum := sha256.Sum256([]byte(value))
	return s.prefix + "fraud:" + string(kind) + ":" + hex.EncodeToString(sum[:12])
}

func (s *Signals) Observe(ctx context.Context, kind model.SignalKind, value, userID string) (int, error) {
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.opTimeout)
	defer cancel()
	key := s.key(kind, value)
	var card *redis.IntCmd
	_, err := s.rdb.TxPipelined(c, func(p redis.Pipeliner) error {
		p.SAdd(c, key, userID)
		p.Expire(c, key, signalWindows[kind])
		card = p.SCard(c, key)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(card.Val()), nil
}
