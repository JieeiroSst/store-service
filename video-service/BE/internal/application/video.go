package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/JIeeiroSst/video-service/internal/domain/model"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var hlsPath = regexp.MustCompile(`^(master\.m3u8|v\d{1,2}/(index\.m3u8|seg_\d{4,6}\.ts))$`)

type VideoService struct {
	repo    port.VideoRepository
	storage port.VideoStorage
	views   port.ViewCounter
	queue   port.JobQueue
}

func NewVideoService(repo port.VideoRepository, storage port.VideoStorage, views port.ViewCounter, queue port.JobQueue) port.VideoUsecase {
	return &VideoService{repo: repo, storage: storage, views: views, queue: queue}
}

func (s *VideoService) Upload(ctx context.Context, in port.UploadInput) (*model.Video, error) {
	if !strings.HasPrefix(in.ContentType, "video/") {
		return nil, fmt.Errorf("%w: content type %q is not a video", port.ErrInvalid, in.ContentType)
	}
	id, err := newID()
	if err != nil {
		return nil, err
	}
	v := model.Video{
		ID:          id,
		Title:       clip(strings.TrimSpace(in.Title), 200),
		Description: clip(strings.TrimSpace(in.Description), 5000),
		ContentType: in.ContentType,
		Status:      model.StatusProcessing,
		CreatedAt:   time.Now().UTC(),
	}

	size, err := s.storage.Put(ctx, v.ObjectKey(), in.Body, -1, v.ContentType)
	if err != nil {
		return nil, fmt.Errorf("store video: %w", err)
	}
	if size == 0 {
		_ = s.storage.Delete(ctx, v.ObjectKey())
		return nil, fmt.Errorf("%w: empty file", port.ErrInvalid)
	}
	v.Size = size

	bg := context.WithoutCancel(ctx)
	if err := s.repo.Save(ctx, v); err != nil {
		_ = s.storage.Delete(bg, v.ObjectKey())
		return nil, fmt.Errorf("save metadata: %w", err)
	}
	if err := s.queue.Enqueue(ctx, v.ID); err != nil {
		_ = s.repo.Delete(bg, v.ID)
		_ = s.storage.Delete(bg, v.ObjectKey())
		return nil, fmt.Errorf("enqueue transcode: %w", err)
	}
	return &v, nil
}

func (s *VideoService) List(ctx context.Context, q port.ListQuery) ([]model.Video, int, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, 0, err
	}
	needle := strings.ToLower(strings.TrimSpace(q.Query))
	var vs []model.Video
	for _, v := range all {
		if v.Status == model.StatusFailed {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(v.Title+" "+v.Description), needle) {
			continue
		}
		vs = append(vs, v)
	}
	if err := s.fillViews(ctx, vs); err != nil {
		return nil, 0, err
	}
	if q.Sort == port.SortPopular {
		sort.SliceStable(vs, func(i, j int) bool { return vs[i].Views > vs[j].Views })
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	pageSize = min(pageSize, maxPageSize)
	page := max(q.Page, 1)
	total := len(vs)
	start := min((page-1)*pageSize, total)
	end := min(start+pageSize, total)
	return vs[start:end], total, nil
}

func (s *VideoService) Related(ctx context.Context, id string, limit int) ([]model.Video, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var vs []model.Video
	for _, v := range all {
		if v.ID != id && v.Status == model.StatusReady {
			vs = append(vs, v)
		}
	}
	if err := s.fillViews(ctx, vs); err != nil {
		return nil, err
	}
	sort.SliceStable(vs, func(i, j int) bool { return vs[i].Views > vs[j].Views })
	if limit <= 0 {
		limit = 12
	}
	return vs[:min(limit, len(vs))], nil
}

func (s *VideoService) Get(ctx context.Context, id string) (*model.Video, error) {
	v, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	counts, err := s.views.Counts(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	v.Views = counts[id]
	return v, nil
}

func (s *VideoService) get(ctx context.Context, id string) (*model.Video, error) {
	if !model.ValidID(id) {
		return nil, port.ErrNotFound
	}
	return s.repo.Get(ctx, id)
}

func (s *VideoService) Open(ctx context.Context, id string) (*model.VideoStream, error) {
	v, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.VideoStream{
		Video:   *v,
		Content: &rangeReader{ctx: ctx, storage: s.storage, key: v.ObjectKey(), size: v.Size},
	}, nil
}

func (s *VideoService) Thumbnail(ctx context.Context, id string) (*model.Asset, error) {
	v, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if v.Status != model.StatusReady {
		return nil, port.ErrNotFound
	}
	data, err := s.storage.Get(ctx, v.ThumbnailKey())
	if err != nil {
		return nil, err
	}
	return &model.Asset{Data: data, ContentType: "image/jpeg"}, nil
}

func (s *VideoService) HLS(ctx context.Context, id, rel string) (*model.Asset, error) {
	if !hlsPath.MatchString(rel) {
		return nil, port.ErrNotFound
	}
	v, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if v.Status != model.StatusReady {
		return nil, port.ErrNotFound
	}
	data, err := s.storage.Get(ctx, v.HLSPrefix()+rel)
	if err != nil {
		return nil, err
	}
	ct := "video/mp2t"
	if strings.HasSuffix(rel, ".m3u8") {
		ct = "application/vnd.apple.mpegurl"
	}
	return &model.Asset{Data: data, ContentType: ct}, nil
}

func (s *VideoService) RecordView(ctx context.Context, id, viewer string) error {
	if _, err := s.get(ctx, id); err != nil {
		return err
	}
	return s.views.Incr(ctx, id, viewer)
}

func (s *VideoService) Delete(ctx context.Context, id string) error {
	v, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, v.ID); err != nil {
		return err
	}
	bg := context.WithoutCancel(ctx)
	err = s.storage.Delete(bg, v.ObjectKey())
	err = errors.Join(err,
		s.storage.Delete(bg, v.ThumbnailKey()),
		s.storage.DeletePrefix(bg, v.HLSPrefix()),
		s.views.Delete(bg, v.ID),
	)
	if err != nil {
		log.Printf("delete %s: leftover objects: %v", v.ID, err)
	}
	return nil
}

func (s *VideoService) fillViews(ctx context.Context, vs []model.Video) error {
	ids := make([]string, len(vs))
	for i, v := range vs {
		ids[i] = v.ID
	}
	counts, err := s.views.Counts(ctx, ids)
	if err != nil {
		return err
	}
	for i := range vs {
		vs[i].Views = counts[vs[i].ID]
	}
	return nil
}

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type rangeReader struct {
	ctx     context.Context
	storage port.VideoStorage
	key     string
	size    int64
	pos     int64
}

func (r *rangeReader) Read(p []byte) (int, error) {
	if r.pos >= r.size {
		return 0, io.EOF
	}
	n, err := r.storage.ReadAt(r.ctx, r.key, r.size, p, r.pos)
	r.pos += int64(n)
	if n > 0 && err == io.EOF {
		err = nil
	}
	return n, err
}

func (r *rangeReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.pos + offset
	case io.SeekEnd:
		abs = r.size + offset
	default:
		return 0, fmt.Errorf("invalid whence %d", whence)
	}
	if abs < 0 {
		return 0, fmt.Errorf("negative position")
	}
	r.pos = abs
	return abs, nil
}
