package application

import (
	"context"
	"fmt"
	"sort"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
)

const (
	minProgress = 0.03
	maxProgress = 0.9

	homeRowSize     = 10
	homeAnchors     = 2
	homeAnchorPool  = 30
	anchorMinWeight = 3
)

type watched struct {
	last     model.Interaction
	watch    *model.Interaction
	progress float64
}

func distinctHistory(history []model.Interaction) []watched {
	sorted := append([]model.Interaction(nil), history...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].At.After(sorted[j].At) })

	index := map[string]int{}
	var out []watched
	for _, in := range sorted {
		i, ok := index[in.VideoID]
		if !ok {
			i = len(out)
			index[in.VideoID] = i
			out = append(out, watched{last: in})
		}
		if in.Type == model.Watch && out[i].watch == nil {
			w := in
			out[i].watch = &w
		}
	}
	return out
}

func slicePage(items []model.Recommendation, q port.PageQuery) *port.Page {
	size := pageSizeOf(q)
	page := max(q.Page, 1)
	start := min((page-1)*size, len(items))
	end := min(start+size, len(items))
	return &port.Page{Items: items[start:end], Total: len(items), Page: page, PageSize: size}
}

func (s *Service) loadHistory(ctx context.Context, userID string) ([]model.Interaction, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user id is required", port.ErrInvalid)
	}
	return s.interactions.ByUser(ctx, userID, historyLimit)
}

func (s *Service) userHistory(ctx context.Context, userID string) ([]watched, error) {
	h, err := s.loadHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	return distinctHistory(h), nil
}

func (s *Service) continueWatching(m *engine.Model, history []watched) []model.Recommendation {
	var out []model.Recommendation
	for _, w := range history {
		if w.watch == nil || w.watch.Value < minProgress || w.watch.Value >= maxProgress {
			continue
		}
		v, ok := m.Video(w.watch.VideoID)
		if !ok {
			continue
		}
		progress, at := w.watch.Value, w.watch.At
		out = append(out, model.Recommendation{Video: v, Reason: model.ReasonContinue, Progress: &progress, At: &at})
	}
	return out
}

func (s *Service) ContinueWatching(ctx context.Context, userID string, q port.PageQuery) (*port.Page, error) {
	history, err := s.userHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	return slicePage(s.continueWatching(s.model.Load(), history), q), nil
}

func (s *Service) History(ctx context.Context, userID string, q port.PageQuery) (*port.Page, error) {
	history, err := s.userHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	m := s.model.Load()
	var out []model.Recommendation
	for _, w := range history {
		v, ok := m.Video(w.last.VideoID)
		if !ok {
			continue
		}
		at := w.last.At
		r := model.Recommendation{Video: v, Reason: model.ReasonHistory, At: &at}
		if w.watch != nil {
			p := w.watch.Value
			r.Progress = &p
		}
		out = append(out, r)
	}
	return slicePage(out, q), nil
}

func (s *Service) RemoveFromHistory(ctx context.Context, userID, videoID string) error {
	if userID == "" || videoID == "" {
		return fmt.Errorf("%w: user id and video id are required", port.ErrInvalid)
	}
	return s.interactions.DeleteByUserVideo(ctx, userID, videoID)
}

func (s *Service) NewReleases(ctx context.Context, q port.PageQuery) (*port.Page, error) {
	return s.rank(ctx, "new", q, func(m *engine.Model) ([]engine.Scored, error) {
		return m.NewReleases(maxCandidates, nil), nil
	})
}

func (s *Service) Home(ctx context.Context, userID string) (*model.Home, error) {
	history, err := s.loadHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	distinct := distinctHistory(history)
	m := s.model.Load()

	seen := map[string]bool{}
	for _, w := range distinct {
		seen[w.last.VideoID] = true
	}

	home := &model.Home{Sections: []model.Section{}}
	shown := map[string]bool{}
	add := func(id, title string, items []model.Recommendation) {
		var row []model.Recommendation
		for _, it := range items {
			if shown[it.Video.ID] {
				continue
			}
			shown[it.Video.ID] = true
			row = append(row, it)
			if len(row) == homeRowSize {
				break
			}
		}
		if len(row) > 0 {
			home.Sections = append(home.Sections, model.Section{ID: id, Title: title, Items: row})
		}
	}

	add("continue_watching", "Continue watching", s.continueWatching(m, distinct))

	forYou := s.hydrate(m, m.ForUser(history, homeRowSize*3))
	personal := false
	for _, r := range forYou {
		if r.Reason != model.ReasonTrending {
			personal = true
			break
		}
	}
	if personal {
		add("for_you", "Recommended for you", forYou)
	}

	anchors := 0
	for _, w := range distinct {
		if anchors == homeAnchors {
			break
		}
		if w.last.Weight() < anchorMinWeight {
			continue
		}
		anchor, ok := m.Video(w.last.VideoID)
		if !ok {
			continue
		}
		similar, _ := m.Similar(anchor.ID, homeAnchorPool)
		var fresh []engine.Scored
		for _, sc := range similar {
			if !seen[sc.VideoID] {
				fresh = append(fresh, sc)
			}
		}
		before := len(home.Sections)
		add("because_"+anchor.ID, "Because you watched "+anchor.Title, s.hydrate(m, fresh))
		if len(home.Sections) > before {
			anchors++
		}
	}

	add("new_releases", "New releases", s.hydrate(m, m.NewReleases(homeAnchorPool, seen)))
	add("trending", "Trending now", s.hydrate(m, m.Trending(homeAnchorPool, seen)))
	return home, nil
}
