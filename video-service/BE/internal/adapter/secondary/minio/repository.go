package minio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/model"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"github.com/minio/minio-go/v7"
)

const metaPrefix = "meta/"

type Repository struct {
	client *minio.Client
	bucket string
}

func NewRepository(client *minio.Client, cfg *config.Config) *Repository {
	return &Repository{client: client, bucket: cfg.Minio.Bucket}
}

func metaKey(id string) string { return metaPrefix + id + ".json" }

func (r *Repository) Save(ctx context.Context, v model.Video) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = r.client.PutObject(ctx, r.bucket, metaKey(v.ID), bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: "application/json"})
	return err
}

func (r *Repository) Get(ctx context.Context, id string) (*model.Video, error) {
	obj, err := r.client.GetObject(ctx, r.bucket, metaKey(id), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return decode(obj, id)
}

func (r *Repository) List(ctx context.Context) ([]model.Video, error) {
	var (
		keys []string
		out  []model.Video
		mu   sync.Mutex
	)
	for o := range r.client.ListObjects(ctx, r.bucket, minio.ListObjectsOptions{Prefix: metaPrefix}) {
		if o.Err != nil {
			return nil, o.Err
		}
		keys = append(keys, o.Key)
	}

	sem := make(chan struct{}, 16)
	var wg sync.WaitGroup
	var firstErr error
	for _, key := range keys {
		wg.Add(1)
		sem <- struct{}{}
		go func(key string) {
			defer wg.Done()
			defer func() { <-sem }()
			id := strings.TrimSuffix(strings.TrimPrefix(key, metaPrefix), ".json")
			v, err := r.Get(ctx, id)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case errors.Is(err, port.ErrNotFound): // deleted while listing
			case err != nil:
				if firstErr == nil {
					firstErr = err
				}
			default:
				out = append(out, *v)
			}
		}(key)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.client.RemoveObject(ctx, r.bucket, metaKey(id), minio.RemoveObjectOptions{})
}

func decode(rd io.Reader, id string) (*model.Video, error) {
	var v model.Video
	if err := json.NewDecoder(rd).Decode(&v); err != nil {
		if err = notFoundOr(err); errors.Is(err, port.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("read metadata %s: %w", id, err)
	}
	return &v, nil
}
