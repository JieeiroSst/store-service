package buffer

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"golang.org/x/sync/singleflight"
)

const (
	loadTimeout     = 30 * time.Second
	maxPrefetchJobs = 8
)

type ChunkStorage struct {
	port.VideoStorage
	chunkSize int64
	prefetch  int
	cache     *lru
	group     singleflight.Group
	jobs      chan struct{}
}

func NewChunkStorage(next port.VideoStorage, cfg *config.Config) *ChunkStorage {
	return &ChunkStorage{
		VideoStorage: next,
		chunkSize:    cfg.Stream.ChunkSize,
		prefetch:     cfg.Stream.PrefetchChunks,
		cache:        newLRU(cfg.Stream.CacheBytes),
		jobs:         make(chan struct{}, maxPrefetchJobs),
	}
}

func (c *ChunkStorage) ReadAt(ctx context.Context, key string, size int64, p []byte, off int64) (int, error) {
	if off >= size {
		return 0, io.EOF
	}
	end := min(off+int64(len(p)), size)

	n := 0
	for pos := off; pos < end; {
		idx := pos / c.chunkSize
		chunk, err := c.chunk(ctx, key, size, idx)
		if err != nil {
			return n, err
		}
		lo := pos - idx*c.chunkSize
		hi := min(int64(len(chunk)), end-idx*c.chunkSize)
		if lo >= hi {
			return n, io.ErrUnexpectedEOF
		}
		n += copy(p[n:], chunk[lo:hi])
		pos = idx*c.chunkSize + hi
	}

	c.prefetchAfter(key, size, (end-1)/c.chunkSize)
	return n, nil
}

func (c *ChunkStorage) chunk(ctx context.Context, key string, size, idx int64) ([]byte, error) {
	ck := chunkKey{key, idx}
	if data, ok := c.cache.get(ck); ok {
		return data, nil
	}

	ch := c.group.DoChan(key+"#"+strconv.FormatInt(idx, 10), func() (any, error) {
		if data, ok := c.cache.get(ck); ok {
			return data, nil
		}
		lctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), loadTimeout)
		defer cancel()
		data, err := c.load(lctx, key, size, idx)
		if err != nil {
			return nil, err
		}
		c.cache.add(ck, data)
		return data, nil
	})

	select {
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]byte), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *ChunkStorage) load(ctx context.Context, key string, size, idx int64) ([]byte, error) {
	off := idx * c.chunkSize
	buf := make([]byte, min(c.chunkSize, size-off))
	n, err := c.VideoStorage.ReadAt(ctx, key, size, buf, off)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("load chunk %d of %s: %w", idx, key, err)
	}
	return buf[:n], nil
}

func (c *ChunkStorage) prefetchAfter(key string, size, last int64) {
	for i := int64(1); i <= int64(c.prefetch); i++ {
		next := last + i
		if next*c.chunkSize >= size || c.cache.contains(chunkKey{key, next}) {
			continue
		}
		select {
		case c.jobs <- struct{}{}:
		default:
			return
		}
		go func(idx int64) {
			defer func() { <-c.jobs }()
			ctx, cancel := context.WithTimeout(context.Background(), loadTimeout)
			defer cancel()
			_, _ = c.chunk(ctx, key, size, idx)
		}(next)
	}
}

func (c *ChunkStorage) Get(ctx context.Context, key string) ([]byte, error) {
	ck := chunkKey{key, -1}
	if data, ok := c.cache.get(ck); ok {
		return data, nil
	}
	ch := c.group.DoChan("obj:"+key, func() (any, error) {
		if data, ok := c.cache.get(ck); ok {
			return data, nil
		}
		lctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), loadTimeout)
		defer cancel()
		data, err := c.VideoStorage.Get(lctx, key)
		if err != nil {
			return nil, err
		}
		c.cache.add(ck, data)
		return data, nil
	})
	select {
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]byte), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
