package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/JIeeiroSst/bot-service/config"
	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
)

type fakeContentUsecase struct {
	calls chan struct{}
}

func (f *fakeContentUsecase) CreatePost(context.Context, port.CreatePostInput) (*model.Post, error) {
	return nil, nil
}
func (f *fakeContentUsecase) UpdatePost(context.Context, string, port.UpdatePostInput) (*model.Post, error) {
	return nil, nil
}
func (f *fakeContentUsecase) GetPost(context.Context, string) (*model.Post, error) { return nil, nil }
func (f *fakeContentUsecase) ListPosts(context.Context, int32, int32, model.PostStatus, string) ([]model.Post, error) {
	return nil, nil
}
func (f *fakeContentUsecase) SubmitForReview(context.Context, string, string) (*model.Post, error) {
	return nil, nil
}
func (f *fakeContentUsecase) ApprovePost(context.Context, string, string) (*model.Post, error) {
	return nil, nil
}
func (f *fakeContentUsecase) RejectPost(context.Context, string, string, string) (*model.Post, error) {
	return nil, nil
}
func (f *fakeContentUsecase) PublishNow(context.Context, string, string) (*model.Post, error) {
	return nil, nil
}
func (f *fakeContentUsecase) CancelScheduledPost(context.Context, string, string) error { return nil }
func (f *fakeContentUsecase) PublishDuePosts(context.Context) error {
	f.calls <- struct{}{}
	return nil
}

func TestSchedulerRunsJob(t *testing.T) {
	cfg := &config.Config{Scheduler: config.SchedulerConfig{PollInterval: "200ms"}}
	fu := &fakeContentUsecase{calls: make(chan struct{}, 1)}
	s := NewScheduler(fu, cfg)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer s.Stop()

	select {
	case <-fu.calls:
	case <-time.After(2 * time.Second):
		t.Fatal("cron job never fired")
	}
}
