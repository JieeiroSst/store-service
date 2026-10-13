package http

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
)

type fakeRecs struct {
	port.RecommendationUsecase
	eventUser string
}

func (f *fakeRecs) RecordEvent(_ context.Context, in model.Interaction) error {
	f.eventUser = in.UserID
	return nil
}

func TestMovieAuthorization(t *testing.T) {
	fakeUserService(t)
	recs := &fakeRecs{}
	cfg := &config.Config{}
	cfg.Server.AllowedOrigins = []string{"*"}
	router := NewRouter(NewHandler(recs), cfg)
	user := testToken("7", "user")
	for _, c := range []struct {
		name string
		call authCall
		want int
	}{
		{"anonymous user feed", authCall{method: "GET", path: "/api/recommendations/users/7"}, 401},
		{"foreign feed", authCall{method: "GET", path: "/api/recommendations/users/8", bearer: user}, 403},
		{"foreign history delete", authCall{method: "DELETE", path: "/api/recommendations/users/8/history/v1", bearer: user}, 403},
		{"user upserts video", authCall{method: "PUT", path: "/api/recommendations/videos/v1", body: `{}`, bearer: user}, 403},
		{"anonymous event", authCall{method: "POST", path: "/api/recommendations/events", body: `{}`}, 401},
	} {
		if got, body := doAuth(router, c.call); got != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, got, c.want, body)
		}
	}
	if code, body := doAuth(router, authCall{method: "POST", path: "/api/recommendations/events", body: `{"user_id":"8","video_id":"v1","type":"view"}`, bearer: user}); code >= 400 || recs.eventUser != "7" {
		t.Fatalf("event owner forced: %d %s user=%s", code, body, recs.eventUser)
	}
}
