package application

import (
	"context"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

type tagService struct {
	trending port.TrendingTagsStore
}

func NewTagService(trending port.TrendingTagsStore) port.TagUsecase {
	return &tagService{trending: trending}
}

func (s *tagService) ListTrending(ctx context.Context, limit int) ([]model.TagCount, error) {
	switch {
	case limit <= 0:
		limit = 10 // smaller default than model.DefaultPageLimit - a trending widget, not a feed
	case limit > model.MaxPageLimit:
		limit = model.MaxPageLimit
	}
	return s.trending.TopTags(ctx, limit)
}
