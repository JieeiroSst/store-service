package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
)

var readScript = redis.NewScript(`
local gen = redis.call('GET', KEYS[1]) or '0'
return {gen, redis.call('GET', ARGV[1] .. gen)}
`)

var writeScript = redis.NewScript(`
local gen = redis.call('GET', KEYS[1]) or '0'
if gen ~= ARGV[1] then return 0 end
redis.call('SET', KEYS[2], ARGV[2], 'PX', ARGV[3])
if gen ~= '0' then redis.call('PEXPIRE', KEYS[1], ARGV[3] * 2) end
return 1
`)

var bumpScript = redis.NewScript(`
local gen = redis.call('INCR', KEYS[1])
redis.call('PEXPIRE', KEYS[1], ARGV[1] * 2)
return gen
`)

type Repository struct {
	port.WishlistRepository
	rdb        *redis.Client
	prefix     string
	ttl        time.Duration
	popularTTL time.Duration
	opTimeout  time.Duration
}

func NewRepository(base port.WishlistRepository, rdb *redis.Client, prefix string, ttl, popularTTL, opTimeout time.Duration) *Repository {
	return &Repository{WishlistRepository: base, rdb: rdb, prefix: prefix, ttl: ttl, popularTTL: popularTTL, opTimeout: opTimeout}
}

func (r *Repository) genKey(id int64) string {
	return r.prefix + "wishlist:gen:" + strconv.FormatInt(id, 10)
}

func (r *Repository) dataPrefix(id int64) string {
	return r.prefix + "wishlist:data:" + strconv.FormatInt(id, 10) + ":"
}

func (r *Repository) keyIndex(key string) string { return r.prefix + "wishlist:key:" + key }

func (r *Repository) op(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), r.opTimeout)
}

func (r *Repository) Get(ctx context.Context, id int64) (*model.Wishlist, error) {
	gen, w := r.load(ctx, id)
	if w != nil {
		return w, nil
	}
	w, err := r.WishlistRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	r.store(ctx, w, gen)
	return w, nil
}

func (r *Repository) GetByKey(ctx context.Context, key string) (*model.Wishlist, error) {
	c, cancel := r.op(ctx)
	idStr, err := r.rdb.Get(c, r.keyIndex(key)).Result()
	cancel()
	if err == nil {
		if id, perr := strconv.ParseInt(idStr, 10, 64); perr == nil {
			if w, gerr := r.Get(ctx, id); gerr == nil && w.Key == key {
				return w, nil
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		warn("get key index", err)
	}
	w, err := r.WishlistRepository.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	c, cancel = r.op(ctx)
	defer cancel()
	if err := r.rdb.Set(c, r.keyIndex(key), w.ID, r.ttl).Err(); err != nil {
		warn("set key index", err)
	}
	return w, nil
}

func (r *Repository) Update(ctx context.Context, w *model.Wishlist) error {
	defer r.invalidate(ctx, w.ID)
	return r.WishlistRepository.Update(ctx, w)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	defer r.invalidate(ctx, id)
	return r.WishlistRepository.Delete(ctx, id)
}

func (r *Repository) AddLine(ctx context.Context, l *model.Line) (*model.Line, error) {
	defer r.invalidate(ctx, l.WishlistID)
	return r.WishlistRepository.AddLine(ctx, l)
}

func (r *Repository) UpdateLineQuantity(ctx context.Context, wishlistID, lineID int64, quantity int) (*model.Line, error) {
	defer r.invalidate(ctx, wishlistID)
	return r.WishlistRepository.UpdateLineQuantity(ctx, wishlistID, lineID, quantity)
}

func (r *Repository) DeleteLine(ctx context.Context, wishlistID, lineID int64) error {
	defer r.invalidate(ctx, wishlistID)
	return r.WishlistRepository.DeleteLine(ctx, wishlistID, lineID)
}

func (r *Repository) ClearLines(ctx context.Context, wishlistID int64) (int64, error) {
	defer r.invalidate(ctx, wishlistID)
	return r.WishlistRepository.ClearLines(ctx, wishlistID)
}

func (r *Repository) MoveLine(ctx context.Context, fromWishlistID, lineID, toWishlistID int64) (*model.Line, error) {
	defer r.invalidate(ctx, fromWishlistID, toWishlistID)
	return r.WishlistRepository.MoveLine(ctx, fromWishlistID, lineID, toWishlistID)
}

func (r *Repository) PopularProducts(ctx context.Context, limit int) ([]model.ProductStat, error) {
	key := r.prefix + "popular:" + strconv.Itoa(limit)
	c, cancel := r.op(ctx)
	raw, err := r.rdb.Get(c, key).Bytes()
	cancel()
	if err == nil {
		var out []model.ProductStat
		if json.Unmarshal(raw, &out) == nil {
			return out, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		warn("get popular products", err)
	}
	out, err := r.WishlistRepository.PopularProducts(ctx, limit)
	if err != nil {
		return nil, err
	}
	if raw, err := json.Marshal(out); err == nil {
		c, cancel := r.op(ctx)
		defer cancel()
		if err := r.rdb.Set(c, key, raw, r.popularTTL).Err(); err != nil {
			warn("set popular products", err)
		}
	}
	return out, nil
}

func (r *Repository) load(ctx context.Context, id int64) (string, *model.Wishlist) {
	c, cancel := r.op(ctx)
	defer cancel()
	res, err := readScript.Run(c, r.rdb, []string{r.genKey(id)}, r.dataPrefix(id)).Slice()
	if err != nil {
		warn("read wishlist", err)
		return "", nil
	}
	gen, _ := res[0].(string)
	raw, ok := res[1].(string)
	if !ok {
		return gen, nil
	}
	var w model.Wishlist
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		return gen, nil
	}
	return gen, &w
}

func (r *Repository) store(ctx context.Context, w *model.Wishlist, gen string) {
	if gen == "" {
		return
	}
	cp := *w
	cp.Owner = nil
	raw, err := json.Marshal(&cp)
	if err != nil {
		return
	}
	c, cancel := r.op(ctx)
	defer cancel()
	err = writeScript.Run(c, r.rdb, []string{r.genKey(w.ID), r.dataPrefix(w.ID) + gen}, gen, raw, r.ttl.Milliseconds()).Err()
	if err != nil {
		warn("write wishlist", err)
	}
}

func (r *Repository) invalidate(ctx context.Context, ids ...int64) {
	c, cancel := r.op(ctx)
	defer cancel()
	for _, id := range ids {
		if err := bumpScript.Run(c, r.rdb, []string{r.genKey(id)}, r.ttl.Milliseconds()).Err(); err != nil {
			warn("invalidate wishlist", err)
		}
	}
}

func warn(op string, err error) {
	slog.Warn("redis cache unavailable", "op", op, "err", err)
}
