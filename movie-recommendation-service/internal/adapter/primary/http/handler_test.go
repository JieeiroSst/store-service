package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
)

type fakeUsecase struct {
	port.RecommendationUsecase
	lastQuery port.PageQuery
	removed   string
	event     model.Interaction
	err       error
}

func (f *fakeUsecase) ForUser(_ context.Context, user string, q port.PageQuery) (*port.Page, error) {
	f.lastQuery = q
	if f.err != nil {
		return nil, f.err
	}
	return &port.Page{
		Items: []model.Recommendation{{Video: model.Video{ID: "b", Title: "B"}, Score: 1, Reason: model.ReasonForYou}},
		Total: 41, Page: q.Page, PageSize: q.PageSize,
	}, nil
}
func (f *fakeUsecase) Similar(context.Context, string, port.PageQuery) (*port.Page, error) {
	return nil, port.ErrNotFound
}
func (f *fakeUsecase) Trending(context.Context, port.PageQuery) (*port.Page, error) {
	return &port.Page{Page: 1, PageSize: 20}, nil
}
func (f *fakeUsecase) RecordEvent(_ context.Context, in model.Interaction) error {
	f.event = in
	return f.err
}
func (f *fakeUsecase) NewReleases(context.Context, port.PageQuery) (*port.Page, error) {
	return &port.Page{Page: 1, PageSize: 20}, nil
}
func (f *fakeUsecase) ContinueWatching(_ context.Context, user string, q port.PageQuery) (*port.Page, error) {
	f.lastQuery = q
	p := 0.5
	return &port.Page{Items: []model.Recommendation{{Video: model.Video{ID: user}, Progress: &p}}, Total: 1, Page: 1, PageSize: 20}, nil
}
func (f *fakeUsecase) History(context.Context, string, port.PageQuery) (*port.Page, error) {
	return &port.Page{Page: 1, PageSize: 20}, f.err
}
func (f *fakeUsecase) RemoveFromHistory(_ context.Context, user, video string) error {
	f.removed = user + "/" + video
	return f.err
}
func (f *fakeUsecase) Home(_ context.Context, user string) (*model.Home, error) {
	return &model.Home{Sections: []model.Section{{ID: "trending", Title: "Trending now"}}}, f.err
}
func (f *fakeUsecase) Ready(context.Context) error { return f.err }

func do(t *testing.T, uc *fakeUsecase, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	cfg := &config.Config{}
	cfg.Server.AllowedOrigins = []string{"*"}
	r := NewRouter(NewHandler(uc), cfg)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
	return w
}

func TestForUser(t *testing.T) {
	uc := &fakeUsecase{}
	w := do(t, uc, "GET", "/api/recommendations/users/u1?page=3&page_size=5&snapshot=tok", "")
	body := w.Body.String()
	if w.Code != 200 || uc.lastQuery != (port.PageQuery{Page: 3, PageSize: 5, Snapshot: "tok"}) ||
		!strings.Contains(body, `"reason":"for_you"`) || !strings.Contains(body, `"total":41`) ||
		!strings.Contains(body, `"page":3`) || !strings.Contains(body, `"page_size":5`) {
		t.Fatalf("%d %s (query %+v)", w.Code, body, uc.lastQuery)
	}
}

func TestTrendingEmptyIsArrayNotNull(t *testing.T) {
	w := do(t, &fakeUsecase{}, "GET", "/api/recommendations/trending", "")
	if w.Code != 200 || !strings.HasPrefix(w.Body.String(), `{"items":[],`) {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
}

func TestSimilarNotFound(t *testing.T) {
	if w := do(t, &fakeUsecase{}, "GET", "/api/recommendations/videos/x/similar", ""); w.Code != http.StatusNotFound {
		t.Fatalf("code = %d", w.Code)
	}
}

func TestRecordEvent(t *testing.T) {
	uc := &fakeUsecase{}
	w := do(t, uc, "POST", "/api/recommendations/events", `{"user_id":"u1","video_id":"v1","type":"rating","value":4}`)
	if w.Code != http.StatusAccepted || uc.event.UserID != "u1" || uc.event.Type != model.Rating || uc.event.Value != 4 {
		t.Fatalf("%d %+v", w.Code, uc.event)
	}
	if w := do(t, uc, "POST", "/api/recommendations/events", `not json`); w.Code != http.StatusBadRequest {
		t.Fatalf("bad json code = %d", w.Code)
	}
	uc.err = port.ErrInvalid
	if w := do(t, uc, "POST", "/api/recommendations/events", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid event code = %d", w.Code)
	}
}

func TestProbes(t *testing.T) {
	if w := do(t, &fakeUsecase{}, "GET", "/healthz", ""); w.Code != 200 {
		t.Fatalf("healthz = %d", w.Code)
	}
	if w := do(t, &fakeUsecase{err: context.DeadlineExceeded}, "GET", "/readyz", ""); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz = %d", w.Code)
	}
}

func TestFeedEndpoints(t *testing.T) {
	uc := &fakeUsecase{}
	if w := do(t, uc, "GET", "/api/recommendations/users/u1/home", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"id":"trending"`) {
		t.Fatalf("home %d %s", w.Code, w.Body)
	}
	w := do(t, uc, "GET", "/api/recommendations/users/u1/continue-watching?page=2", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"progress":0.5`) || uc.lastQuery.Page != 2 {
		t.Fatalf("continue-watching %d %s", w.Code, w.Body)
	}
	if w := do(t, uc, "GET", "/api/recommendations/users/u1/history", ""); w.Code != 200 || !strings.HasPrefix(w.Body.String(), `{"items":[],`) {
		t.Fatalf("history %d %s", w.Code, w.Body)
	}
	if w := do(t, uc, "GET", "/api/recommendations/new", ""); w.Code != 200 {
		t.Fatalf("new %d", w.Code)
	}
	if w := do(t, uc, "DELETE", "/api/recommendations/users/u1/history/v9", ""); w.Code != http.StatusNoContent || uc.removed != "u1/v9" {
		t.Fatalf("remove %d %q", w.Code, uc.removed)
	}
	uc.err = port.ErrInvalid
	if w := do(t, uc, "DELETE", "/api/recommendations/users/u1/history/v9", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("remove invalid %d", w.Code)
	}
}
