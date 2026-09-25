package application

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
)

const (
	sourceSync   = "sync"
	sourceManual = "manual"

	historyLimit = 200

	defaultPageSize = 20
	maxPageSize     = 50
	maxCandidates   = 500
)

type Service struct {
	videos       port.VideoRepository
	interactions port.InteractionRepository
	source       port.CatalogSource
	health       port.HealthChecker
	cfg          *config.Config
	now          func() time.Time
	snapshots    port.SnapshotRepository
	model        atomic.Pointer[engine.Model]
}

var (
	_ port.RecommendationUsecase = (*Service)(nil)
	_ port.ModelMaintainer       = (*Service)(nil)
)

func NewService(
	videos port.VideoRepository,
	interactions port.InteractionRepository,
	snapshots port.SnapshotRepository,
	source port.CatalogSource,
	health port.HealthChecker,
	cfg *config.Config,
) *Service {
	s := &Service{videos: videos, interactions: interactions, snapshots: snapshots, source: source, health: health, cfg: cfg, now: time.Now}
	s.model.Store(engine.Build(nil, nil, s.now(), cfg.Engine))
	return s
}

func pageSizeOf(q port.PageQuery) int {
	if q.PageSize <= 0 {
		return defaultPageSize
	}
	return min(q.PageSize, maxPageSize)
}

func (s *Service) paginate(m *engine.Model, ranked []engine.Scored, q port.PageQuery, token string) *port.Page {
	size := pageSizeOf(q)
	page := max(q.Page, 1)
	total := len(ranked)
	start := min((page-1)*size, total)
	end := min(start+size, total)
	return &port.Page{
		Items:    s.hydrate(m, ranked[start:end]),
		Total:    total,
		Page:     page,
		PageSize: size,
		Snapshot: token,
	}
}

func (s *Service) ForUser(ctx context.Context, userID string, q port.PageQuery) (*port.Page, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user id is required", port.ErrInvalid)
	}
	return s.rank(ctx, "user:"+userID, q, func(m *engine.Model) ([]engine.Scored, error) {
		history, err := s.interactions.ByUser(ctx, userID, historyLimit)
		if err != nil {
			return nil, err
		}
		return m.ForUser(history, maxCandidates), nil
	})
}

func (s *Service) Similar(ctx context.Context, videoID string, q port.PageQuery) (*port.Page, error) {
	return s.rank(ctx, "similar:"+videoID, q, func(m *engine.Model) ([]engine.Scored, error) {
		ranked, ok := m.Similar(videoID, maxCandidates)
		if !ok {
			return nil, port.ErrNotFound
		}
		return ranked, nil
	})
}

func (s *Service) Trending(ctx context.Context, q port.PageQuery) (*port.Page, error) {
	return s.rank(ctx, "trending", q, func(m *engine.Model) ([]engine.Scored, error) {
		return m.Trending(maxCandidates, nil), nil
	})
}

func (s *Service) Cleanup(ctx context.Context) error { return s.snapshots.DeleteExpired(ctx, s.now()) }

func (s *Service) hydrate(m *engine.Model, scored []engine.Scored) []model.Recommendation {
	out := make([]model.Recommendation, 0, len(scored))
	for _, sc := range scored {
		v, ok := m.Video(sc.VideoID)
		if !ok {
			continue
		}
		r := model.Recommendation{Video: v, Score: sc.Score, Reason: sc.Reason}
		if sc.BecauseID != "" {
			if b, ok := m.Video(sc.BecauseID); ok {
				r.Because = &b
			}
		}
		out = append(out, r)
	}
	return out
}

func (s *Service) RecordEvent(ctx context.Context, in model.Interaction) error {
	if err := in.Validate(); err != nil {
		return fmt.Errorf("%w: %v", port.ErrInvalid, err)
	}
	if in.At.IsZero() {
		in.At = s.now()
	}
	return s.interactions.Add(ctx, in)
}

func (s *Service) UpsertVideo(ctx context.Context, v model.Video) error {
	if v.ID == "" || v.Title == "" {
		return fmt.Errorf("%w: id and title are required", port.ErrInvalid)
	}
	if v.Status == "" {
		v.Status = model.StatusReady
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = s.now()
	}
	if err := s.videos.Upsert(ctx, sourceManual, []model.Video{v}); err != nil {
		return err
	}
	return s.Train(ctx)
}

func (s *Service) DeleteVideo(ctx context.Context, id string) error {
	if err := s.videos.Delete(ctx, id); err != nil {
		return err
	}
	if err := s.interactions.DeleteByVideo(ctx, id); err != nil {
		return err
	}
	return s.Train(ctx)
}

func (s *Service) Ready(ctx context.Context) error { return s.health.Ping(ctx) }

func (s *Service) Sync(ctx context.Context) error {
	fetched, err := s.source.Fetch(ctx)
	if err == nil {
		if err = s.videos.Upsert(ctx, sourceSync, fetched); err == nil {
			keep := make([]string, len(fetched))
			for i, v := range fetched {
				keep[i] = v.ID
			}
			err = s.videos.DeleteMissing(ctx, sourceSync, keep)
		}
	}
	if terr := s.Train(ctx); terr != nil {
		return terr
	}
	if err != nil {
		return fmt.Errorf("sync catalog: %w", err)
	}
	return nil
}

func (s *Service) Train(ctx context.Context) error {
	began := time.Now()
	start := s.now()
	videos, err := s.videos.List(ctx)
	if err != nil {
		return fmt.Errorf("load videos: %w", err)
	}
	interactions, err := s.interactions.Since(ctx, start.Add(-s.cfg.Train.Window))
	if err != nil {
		return fmt.Errorf("load interactions: %w", err)
	}
	m := engine.Build(videos, interactions, start, s.cfg.Engine)
	s.model.Store(m)
	log.Printf("model trained: %d videos, %d interactions in %s", m.Len(), len(interactions), time.Since(began).Round(time.Millisecond))
	return nil
}
