package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"github.com/JIeeiroSst/wishlists-service/pkg/credential"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const missingMarker = "-"

type Users struct {
	base       port.UserDirectory
	rdb        *redis.Client
	prefix     string
	ttl        time.Duration
	missingTTL time.Duration
	opTimeout  time.Duration
	group      singleflight.Group
}

func NewUsers(base port.UserDirectory, rdb *redis.Client, prefix string, ttl, missingTTL, opTimeout time.Duration) *Users {
	return &Users{base: base, rdb: rdb, prefix: prefix, ttl: ttl, missingTTL: missingTTL, opTimeout: opTimeout}
}

func (u *Users) GetUser(ctx context.Context, id string) (*model.Owner, error) {
	key := u.prefix + "user:" + id
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), u.opTimeout)
	raw, err := u.rdb.Get(c, key).Result()
	cancel()
	switch {
	case err == nil && raw == missingMarker:
		return nil, model.ErrNotFound
	case err == nil:
		var o model.Owner
		if json.Unmarshal([]byte(raw), &o) == nil {
			return &o, nil
		}
	case !errors.Is(err, redis.Nil):
		warn("get user", err)
	}
	if credential.From(ctx).Empty() {
		return nil, model.ErrUnauthenticated
	}
	v, err, _ := u.group.Do(id, func() (any, error) {
		o, err := u.base.GetUser(context.WithoutCancel(ctx), id)
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), u.opTimeout)
		defer cancel()
		switch {
		case errors.Is(err, model.ErrNotFound):
			if serr := u.rdb.Set(c, key, missingMarker, u.missingTTL).Err(); serr != nil {
				warn("set missing user", serr)
			}
		case err == nil:
			if raw, merr := json.Marshal(o); merr == nil {
				if serr := u.rdb.Set(c, key, raw, u.ttl).Err(); serr != nil {
					warn("set user", serr)
				}
			}
		}
		return o, err
	})
	if err != nil {
		return nil, err
	}
	o := *v.(*model.Owner)
	return &o, nil
}
