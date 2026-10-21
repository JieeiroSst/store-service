package application

import (
	"context"
	"testing"

	"aiagentsystem/internal/domain/model"
)

func TestHistoryListLimits(t *testing.T) {
	cases := []struct {
		name  string
		limit int
		want  int
	}{
		{"default", 0, DefaultHistoryLimit},
		{"explicit", 10, 10},
		{"capped", 1000, MaxHistoryLimit},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeHistory{}
			if _, err := NewHistoryService(repo).List(context.Background(), "alice", tc.limit); err != nil {
				t.Fatalf("List: %v", err)
			}
			if repo.gotLimit != tc.want || repo.gotUser != "alice" {
				t.Fatalf("repo got (%q, %d), want (alice, %d)", repo.gotUser, repo.gotLimit, tc.want)
			}
		})
	}
}

func TestHistoryListRejectsInvalidInput(t *testing.T) {
	svc := NewHistoryService(&fakeHistory{})

	if _, err := svc.List(context.Background(), "", 10); model.AsError(err).Kind != model.KindBadRequest {
		t.Fatalf("empty user: err = %v", err)
	}
	if _, err := svc.List(context.Background(), "alice", -1); model.AsError(err).Kind != model.KindBadRequest {
		t.Fatalf("negative limit: err = %v", err)
	}
}
