package mfastore

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "user_service:mfa:challenge:"

type Store struct {
	rdb redis.UniversalClient
}

func New(rdb redis.UniversalClient) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) Save(ctx context.Context, token string, challenge domain.MFAChallenge, ttl time.Duration) error {
	payload, err := json.Marshal(challenge)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, keyPrefix+token, payload, ttl).Err()
}

func (s *Store) Get(ctx context.Context, token string) (domain.MFAChallenge, error) {
	raw, err := s.rdb.Get(ctx, keyPrefix+token).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.MFAChallenge{}, domain.ErrMFAChallenge
	}
	if err != nil {
		return domain.MFAChallenge{}, err
	}
	var c domain.MFAChallenge
	if err := json.Unmarshal(raw, &c); err != nil {
		return domain.MFAChallenge{}, domain.ErrMFAChallenge
	}
	return c, nil
}

func (s *Store) Delete(ctx context.Context, token string) (bool, error) {
	n, err := s.rdb.Del(ctx, keyPrefix+token).Result()
	return n == 1, err
}
