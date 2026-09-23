package buffer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/model"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"golang.org/x/sync/singleflight"
)

type CachedRepository struct {
	next    port.VideoRepository
	bus     port.InvalidationBus
	metaTTL time.Duration
	listTTL time.Duration
	group   singleflight.Group

	mu       sync.RWMutex
	byID     map[string]metaEntry
	list     []model.Video
	listedAt time.Time
}

type metaEntry struct {
	video   model.Video
	expires time.Time
}

func NewCachedRepository(next port.VideoRepository, bus port.InvalidationBus, cfg *config.Config) *CachedRepository {
	return &CachedRepository{
		next:    next,
		bus:     bus,
		metaTTL: cfg.Stream.MetaTTL,
		listTTL: cfg.Stream.ListTTL,
		byID:    map[string]metaEntry{},
	}
}

func (r *CachedRepository) Save(ctx context.Context, v model.Video) error {
	if err := r.next.Save(ctx, v); err != nil {
		return err
	}
	r.changed(ctx, v.ID)
	return nil
}

func (r *CachedRepository) Delete(ctx context.Context, id string) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	r.changed(ctx, id)
	return nil
}

func (r *CachedRepository) Get(ctx context.Context, id string) (*model.Video, error) {
	r.mu.RLock()
	e, ok := r.byID[id]
	r.mu.RUnlock()
	if ok && time.Now().Before(e.expires) {
		v := e.video
		return &v, nil
	}

	res, err, _ := r.group.Do("get:"+id, func() (any, error) {
		return r.next.Get(context.WithoutCancel(ctx), id)
	})
	if err != nil {
		return nil, err
	}
	v := *res.(*model.Video)
	r.mu.Lock()
	r.byID[id] = metaEntry{video: v, expires: time.Now().Add(r.metaTTL)}
	r.mu.Unlock()
	return &v, nil
}

func (r *CachedRepository) List(ctx context.Context) ([]model.Video, error) {
	r.mu.RLock()
	list, fresh := r.list, r.list != nil && time.Since(r.listedAt) < r.listTTL
	r.mu.RUnlock()
	if fresh {
		return list, nil
	}

	res, err, _ := r.group.Do("list", func() (any, error) {
		return r.next.List(context.WithoutCancel(ctx))
	})
	if err != nil {
		return nil, err
	}
	list = res.([]model.Video)
	r.mu.Lock()
	r.list, r.listedAt = list, time.Now()
	r.mu.Unlock()
	return list, nil
}

func (r *CachedRepository) changed(ctx context.Context, id string) {
	r.invalidate(id)
	if r.bus != nil {
		if err := r.bus.Publish(context.WithoutCancel(ctx), id); err != nil {
			log.Printf("publish invalidation for %s: %v", id, err)
		}
	}
}

func (r *CachedRepository) Listen(ctx context.Context) {
	if r.bus == nil {
		return
	}
	for ctx.Err() == nil {
		if err := r.bus.Subscribe(ctx, r.invalidate); err != nil && ctx.Err() == nil {
			log.Printf("invalidation subscription: %v", err)
			time.Sleep(2 * time.Second)
		}
	}
}

func (r *CachedRepository) invalidate(id string) {
	r.mu.Lock()
	delete(r.byID, id)
	r.list = nil
	r.mu.Unlock()
}
