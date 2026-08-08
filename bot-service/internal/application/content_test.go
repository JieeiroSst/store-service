package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
)

type fakePublisher struct {
	ch    model.Channel
	fail  bool
	calls int
}

func (f *fakePublisher) Channel() model.Channel { return f.ch }
func (f *fakePublisher) Publish(ctx context.Context, post model.Post) (string, error) {
	f.calls++
	if f.fail {
		return "", errPublishFailed
	}
	return "ext-id", nil
}

var errPublishFailed = &publishError{}

type publishError struct{}

func (*publishError) Error() string { return "publish failed" }

// fakeRepo is a minimal in-memory port.PostRepository: enough for GetByID to
// actually round-trip what Create/Update stored, which the workflow methods
// (SubmitForReview, ApprovePost, ...) depend on.
type fakeRepo struct {
	posts   map[string]*model.Post
	created []model.Post
	updated []model.Post
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{posts: make(map[string]*model.Post)}
}

func (f *fakeRepo) Create(ctx context.Context, post *model.Post) error {
	f.created = append(f.created, *post)
	cp := *post
	f.posts[post.ID] = &cp
	return nil
}
func (f *fakeRepo) Update(ctx context.Context, post *model.Post) error {
	f.updated = append(f.updated, *post)
	cp := *post
	f.posts[post.ID] = &cp
	return nil
}
func (f *fakeRepo) GetByID(ctx context.Context, id string) (*model.Post, error) {
	p, ok := f.posts[id]
	if !ok {
		return nil, fmt.Errorf("post %q not found", id)
	}
	cp := *p
	return &cp, nil
}
func (f *fakeRepo) List(context.Context, int32, int32, model.PostStatus, string) ([]model.Post, error) {
	return nil, nil
}

// DueForPublish/ActiveRecurring deliberately do NOT filter by status here
// (unlike the real gorm-backed repository) - they return every post with a
// past ScheduledAt/a CronExpr, so tests can exercise contentService's own
// defense-in-depth status guard in PublishDuePosts, not just the repo query.
func (f *fakeRepo) DueForPublish(ctx context.Context, before time.Time) ([]model.Post, error) {
	var out []model.Post
	for _, p := range f.posts {
		if p.CronExpr == "" && !p.ScheduledAt.IsZero() && !p.ScheduledAt.After(before) {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (f *fakeRepo) ActiveRecurring(context.Context) ([]model.Post, error) {
	var out []model.Post
	for _, p := range f.posts {
		if p.CronExpr != "" {
			out = append(out, *p)
		}
	}
	return out, nil
}

// seed inserts a post directly, bypassing CreatePost's validation, for tests
// that need to start from an arbitrary status.
func (f *fakeRepo) seed(post model.Post) {
	cp := post
	f.posts[post.ID] = &cp
}

func newTestService(fail bool) (*contentService, *fakeRepo, *fakePublisher) {
	repo := newFakeRepo()
	pub := &fakePublisher{ch: model.ChannelTelegram, fail: fail}
	registry := PublisherRegistry{model.ChannelTelegram: pub}
	return &contentService{repo: repo, publishers: registry}, repo, pub
}

func TestRunRecurring_IncrementsCountersOnSuccess(t *testing.T) {
	s, _, _ := newTestService(false)
	now := time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)

	post := &model.Post{
		ID:       "p1",
		Channels: model.ChannelList{model.ChannelTelegram},
		CronExpr: "@every 1h",
		Status:   model.PostStatusActive,
	}

	if err := s.runRecurring(context.Background(), post, now, false, "scheduler"); err != nil {
		t.Fatalf("runRecurring: %v", err)
	}

	if post.RunsToday != 1 || post.RunsThisMonth != 1 {
		t.Fatalf("expected counters at 1, got today=%d month=%d", post.RunsToday, post.RunsThisMonth)
	}
	if post.LastRunDate != "2026-03-05" || post.LastRunMonth != "2026-03" {
		t.Fatalf("unexpected LastRunDate/LastRunMonth: %q %q", post.LastRunDate, post.LastRunMonth)
	}
	if post.Status != model.PostStatusActive {
		t.Fatalf("recurring post should stay active, got %q", post.Status)
	}
	if len(post.Results) != 1 || post.Results[0].Error != "" {
		t.Fatalf("expected a clean result, got %+v", post.Results)
	}
}

func TestRunRecurring_StaysActiveOnChannelFailure(t *testing.T) {
	s, _, _ := newTestService(true)
	now := time.Now()

	post := &model.Post{
		ID:       "p1",
		Channels: model.ChannelList{model.ChannelTelegram},
		CronExpr: "@every 1h",
		Status:   model.PostStatusActive,
	}

	if err := s.runRecurring(context.Background(), post, now, false, "scheduler"); err != nil {
		t.Fatalf("runRecurring: %v", err)
	}

	if post.Status != model.PostStatusActive {
		t.Fatalf("a failed run must not kill a recurring post, got status %q", post.Status)
	}
	if post.RunsToday != 1 {
		t.Fatalf("a failed run still counts as an attempt toward the cap, got RunsToday=%d", post.RunsToday)
	}
	if post.Results[0].Error == "" {
		t.Fatalf("expected the failure to be recorded in Results")
	}
}

func TestRunRecurring_SkipsWhenDailyCapReached(t *testing.T) {
	s, _, pub := newTestService(false)
	now := time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)

	post := &model.Post{
		ID:            "p1",
		Channels:      model.ChannelList{model.ChannelTelegram},
		CronExpr:      "@every 1h",
		Status:        model.PostStatusActive,
		MaxRunsPerDay: 2,
		RunsToday:     2,
		LastRunDate:   "2026-03-05",
	}

	if err := s.runRecurring(context.Background(), post, now, false, "scheduler"); err != nil {
		t.Fatalf("runRecurring: %v", err)
	}

	if post.RunsToday != 2 {
		t.Fatalf("capped run must not increment the counter further, got %d", post.RunsToday)
	}
	if post.LastRunAt == nil || !post.LastRunAt.Equal(now) {
		t.Fatalf("capped run must still advance LastRunAt so the scheduler doesn't retry forever")
	}
	if len(post.Results) != 1 || post.Results[0].Error == "" {
		t.Fatalf("expected a skipped-cap result, got %+v", post.Results)
	}
	if pub.calls != 0 {
		t.Fatalf("a capped run must not call the publisher at all, got %d calls", pub.calls)
	}
}

func TestRunRecurring_ResetsCountersOnDayRollover(t *testing.T) {
	s, _, _ := newTestService(false)
	now := time.Date(2026, 3, 6, 0, 5, 0, 0, time.UTC) // next day

	post := &model.Post{
		ID:            "p1",
		Channels:      model.ChannelList{model.ChannelTelegram},
		CronExpr:      "@every 1h",
		Status:        model.PostStatusActive,
		MaxRunsPerDay: 2,
		RunsToday:     2, // was capped yesterday
		LastRunDate:   "2026-03-05",
		RunsThisMonth: 40,
		LastRunMonth:  "2026-03",
	}

	if err := s.runRecurring(context.Background(), post, now, false, "scheduler"); err != nil {
		t.Fatalf("runRecurring: %v", err)
	}

	if post.RunsToday != 1 {
		t.Fatalf("expected RunsToday to reset then increment to 1 on day rollover, got %d", post.RunsToday)
	}
	if post.RunsThisMonth != 41 {
		t.Fatalf("month counter should carry over within the same month, got %d", post.RunsThisMonth)
	}
	if post.LastRunDate != "2026-03-06" {
		t.Fatalf("expected LastRunDate to advance to the new day, got %q", post.LastRunDate)
	}
}

func TestRunRecurring_LocksAsPublishingDuringDispatch(t *testing.T) {
	s, repo, _ := newTestService(false)
	post := &model.Post{
		ID:       "p1",
		Channels: model.ChannelList{model.ChannelTelegram},
		CronExpr: "@every 1h",
		Status:   model.PostStatusActive,
	}

	if err := s.runRecurring(context.Background(), post, time.Now(), false, "scheduler"); err != nil {
		t.Fatalf("runRecurring: %v", err)
	}

	if len(repo.updated) != 2 {
		t.Fatalf("expected two persisted updates (lock + final), got %d", len(repo.updated))
	}
	if repo.updated[0].Status != model.PostStatusPublishing {
		t.Fatalf("expected the first persisted update to lock the post as publishing, got %q", repo.updated[0].Status)
	}
	if repo.updated[1].Status != model.PostStatusActive {
		t.Fatalf("expected the post to return to active after dispatch, got %q", repo.updated[1].Status)
	}
}

func TestPublishOneOff_SetsTerminalStatus(t *testing.T) {
	s, _, _ := newTestService(false)
	post := &model.Post{ID: "p1", Channels: model.ChannelList{model.ChannelTelegram}, Status: model.PostStatusScheduled}

	if err := s.publishOneOff(context.Background(), post, "scheduler"); err != nil {
		t.Fatalf("publishOneOff: %v", err)
	}

	if post.Status != model.PostStatusPublished {
		t.Fatalf("expected published, got %q", post.Status)
	}
}

func TestPublishOneOff_FailedChannelMarksPostFailed(t *testing.T) {
	s, _, _ := newTestService(true)
	post := &model.Post{ID: "p1", Channels: model.ChannelList{model.ChannelTelegram}, Status: model.PostStatusScheduled}

	if err := s.publishOneOff(context.Background(), post, "scheduler"); err != nil {
		t.Fatalf("publishOneOff: %v", err)
	}

	if post.Status != model.PostStatusFailed {
		t.Fatalf("expected failed, got %q", post.Status)
	}
}

func TestPublishOneOff_LocksAsPublishingDuringDispatch(t *testing.T) {
	s, repo, _ := newTestService(false)
	post := &model.Post{ID: "p1", Channels: model.ChannelList{model.ChannelTelegram}, Status: model.PostStatusScheduled}

	if err := s.publishOneOff(context.Background(), post, "alice"); err != nil {
		t.Fatalf("publishOneOff: %v", err)
	}

	if len(repo.updated) != 2 {
		t.Fatalf("expected two persisted updates (lock + final), got %d", len(repo.updated))
	}
	if repo.updated[0].Status != model.PostStatusPublishing {
		t.Fatalf("expected the first persisted update to lock the post as publishing, got %q", repo.updated[0].Status)
	}
	if repo.updated[0].StatusChangedBy != "alice" {
		t.Fatalf("expected StatusChangedBy to be recorded, got %q", repo.updated[0].StatusChangedBy)
	}
}

// Retrying a partially-failed one-off post must not re-publish to a channel
// that already succeeded - that would post a duplicate.
func TestPublishOneOff_DoesNotRePublishAlreadySucceededChannel(t *testing.T) {
	s, _, pub := newTestService(false)
	post := &model.Post{
		ID:       "p1",
		Channels: model.ChannelList{model.ChannelTelegram},
		Status:   model.PostStatusFailed,
		Results: model.ResultList{
			{Channel: model.ChannelTelegram, ExternalID: "already-sent", PublishedAt: time.Now()},
		},
	}

	if err := s.publishOneOff(context.Background(), post, "alice"); err != nil {
		t.Fatalf("publishOneOff: %v", err)
	}

	if pub.calls != 0 {
		t.Fatalf("expected the publisher not to be called again, got %d calls", pub.calls)
	}
	if post.Results[0].ExternalID != "already-sent" {
		t.Fatalf("expected the prior successful result to be carried forward, got %+v", post.Results[0])
	}
	if post.Status != model.PostStatusPublished {
		t.Fatalf("expected published (the only channel already succeeded), got %q", post.Status)
	}
}

// A post may carry text only, media only, or both - it just can't carry neither.
func TestCreatePost_AllowsTextOnly(t *testing.T) {
	s, repo, _ := newTestService(false)

	post, err := s.CreatePost(context.Background(), port.CreatePostInput{
		Text:     "hello",
		Channels: []model.Channel{model.ChannelTelegram},
	})
	if err != nil {
		t.Fatalf("text-only post should be allowed, got error: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected the post to be persisted")
	}
	if post.MediaKind != model.MediaKindTextOnly {
		t.Fatalf("expected derived media_kind text_only, got %q", post.MediaKind)
	}
}

func TestCreatePost_AllowsMediaOnly(t *testing.T) {
	s, repo, _ := newTestService(false)

	post, err := s.CreatePost(context.Background(), port.CreatePostInput{
		Media:    []model.Media{{URL: "https://example.com/a.jpg", Type: model.MediaTypeImage}},
		Channels: []model.Channel{model.ChannelTelegram},
	})
	if err != nil {
		t.Fatalf("media-only post should be allowed, got error: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected the post to be persisted")
	}
	if post.MediaKind != model.MediaKindSingleImage {
		t.Fatalf("expected derived media_kind single_image, got %q", post.MediaKind)
	}
}

func TestCreatePost_DerivesMultiImageAndVideoKinds(t *testing.T) {
	s, _, _ := newTestService(false)

	multi, err := s.CreatePost(context.Background(), port.CreatePostInput{
		Media: []model.Media{
			{URL: "https://example.com/a.jpg", Type: model.MediaTypeImage},
			{URL: "https://example.com/b.jpg", Type: model.MediaTypeImage},
		},
		Channels: []model.Channel{model.ChannelTelegram},
	})
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if multi.MediaKind != model.MediaKindMultiImage {
		t.Fatalf("expected multi_image, got %q", multi.MediaKind)
	}

	video, err := s.CreatePost(context.Background(), port.CreatePostInput{
		Media:    []model.Media{{URL: "https://example.com/a.mp4", Type: model.MediaTypeVideo}},
		Channels: []model.Channel{model.ChannelTelegram},
	})
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if video.MediaKind != model.MediaKindVideo {
		t.Fatalf("expected video, got %q", video.MediaKind)
	}
}

func TestCreatePost_ExplicitMediaKindOverridesDerivation(t *testing.T) {
	s, _, _ := newTestService(false)

	post, err := s.CreatePost(context.Background(), port.CreatePostInput{
		Media:     []model.Media{{URL: "https://example.com/a.mp4", Type: model.MediaTypeVideo}},
		MediaKind: model.MediaKindReel,
		Channels:  []model.Channel{model.ChannelTelegram},
	})
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post.MediaKind != model.MediaKindReel {
		t.Fatalf("expected explicit reel to win over derivation, got %q", post.MediaKind)
	}
}

func TestCreatePost_AllowsTextAndMedia(t *testing.T) {
	s, repo, _ := newTestService(false)

	_, err := s.CreatePost(context.Background(), port.CreatePostInput{
		Text:     "caption",
		Media:    []model.Media{{URL: "https://example.com/a.jpg", Type: model.MediaTypeImage}},
		Channels: []model.Channel{model.ChannelTelegram},
	})
	if err != nil {
		t.Fatalf("text+media post should be allowed, got error: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected the post to be persisted")
	}
}

func TestCreatePost_RejectsNeitherTextNorMedia(t *testing.T) {
	s, repo, _ := newTestService(false)

	_, err := s.CreatePost(context.Background(), port.CreatePostInput{
		Channels: []model.Channel{model.ChannelTelegram},
	})
	if err == nil {
		t.Fatal("expected an error when a post has neither text nor media")
	}
	if len(repo.created) != 0 {
		t.Fatalf("post with no content must not be persisted")
	}
}

func TestCreatePost_DraftSkipsChannelRequirement(t *testing.T) {
	s, repo, _ := newTestService(false)

	post, err := s.CreatePost(context.Background(), port.CreatePostInput{
		SaveAsDraft: true,
		Title:       "Untitled campaign",
		Text:        "still figuring out the channels",
		CreatedBy:   "alice",
	})
	if err != nil {
		t.Fatalf("a draft with no channels should be allowed, got error: %v", err)
	}
	if post.Status != model.PostStatusDraft {
		t.Fatalf("expected draft status, got %q", post.Status)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected the draft to be persisted")
	}
	if post.CreatedBy != "alice" || post.StatusChangedBy != "alice" {
		t.Fatalf("expected CreatedBy/StatusChangedBy to be recorded, got %q/%q", post.CreatedBy, post.StatusChangedBy)
	}
}

func TestUpdatePost_OnlyAllowedForDraftOrPendingReview(t *testing.T) {
	s, repo, _ := newTestService(false)
	repo.seed(model.Post{ID: "p1", Text: "x", Status: model.PostStatusPublished})

	newText := "y"
	_, err := s.UpdatePost(context.Background(), "p1", port.UpdatePostInput{Text: &newText})
	if err == nil {
		t.Fatal("expected editing a published post to be rejected")
	}
}

func TestUpdatePost_DemotesPendingReviewToDraft(t *testing.T) {
	s, repo, _ := newTestService(false)
	repo.seed(model.Post{
		ID:       "p1",
		Text:     "original",
		Channels: model.ChannelList{model.ChannelTelegram},
		Status:   model.PostStatusPendingReview,
	})

	newText := "revised"
	post, err := s.UpdatePost(context.Background(), "p1", port.UpdatePostInput{Text: &newText})
	if err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}
	if post.Text != "revised" {
		t.Fatalf("expected text to be updated, got %q", post.Text)
	}
	if post.Status != model.PostStatusDraft {
		t.Fatalf("editing a pending-review post must demote it back to draft, got %q", post.Status)
	}
}

func TestSubmitForReview_RequiresChannelsAndSchedule(t *testing.T) {
	s, repo, _ := newTestService(false)
	repo.seed(model.Post{ID: "p1", Text: "x", Status: model.PostStatusDraft})

	if _, err := s.SubmitForReview(context.Background(), "p1", "alice"); err == nil {
		t.Fatal("expected submit to fail without channels or a schedule")
	}
}

func TestSubmitForReview_MovesDraftToPendingReview(t *testing.T) {
	s, repo, _ := newTestService(false)
	repo.seed(model.Post{
		ID:          "p1",
		Text:        "x",
		Channels:    model.ChannelList{model.ChannelTelegram},
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      model.PostStatusDraft,
	})

	post, err := s.SubmitForReview(context.Background(), "p1", "alice")
	if err != nil {
		t.Fatalf("SubmitForReview: %v", err)
	}
	if post.Status != model.PostStatusPendingReview {
		t.Fatalf("expected pending_review, got %q", post.Status)
	}
	if post.StatusChangedBy != "alice" {
		t.Fatalf("expected StatusChangedBy to be recorded, got %q", post.StatusChangedBy)
	}
}

func TestApprovePost_MovesToScheduled(t *testing.T) {
	s, repo, _ := newTestService(false)
	repo.seed(model.Post{
		ID:          "p1",
		Text:        "x",
		Channels:    model.ChannelList{model.ChannelTelegram},
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      model.PostStatusPendingReview,
	})

	post, err := s.ApprovePost(context.Background(), "p1", "bob")
	if err != nil {
		t.Fatalf("ApprovePost: %v", err)
	}
	if post.Status != model.PostStatusScheduled {
		t.Fatalf("expected scheduled, got %q", post.Status)
	}
	if post.ApprovedBy != "bob" || post.StatusChangedBy != "bob" {
		t.Fatalf("expected ApprovedBy/StatusChangedBy to be recorded, got %q/%q", post.ApprovedBy, post.StatusChangedBy)
	}
}

func TestApprovePost_RecurringMovesToActive(t *testing.T) {
	s, repo, _ := newTestService(false)
	repo.seed(model.Post{
		ID:       "p1",
		Text:     "x",
		Channels: model.ChannelList{model.ChannelTelegram},
		CronExpr: "@every 1h",
		Status:   model.PostStatusPendingReview,
	})

	post, err := s.ApprovePost(context.Background(), "p1", "bob")
	if err != nil {
		t.Fatalf("ApprovePost: %v", err)
	}
	if post.Status != model.PostStatusActive {
		t.Fatalf("expected active for a recurring post, got %q", post.Status)
	}
}

func TestRejectPost_SendsBackToDraftWithReason(t *testing.T) {
	s, repo, _ := newTestService(false)
	repo.seed(model.Post{
		ID:       "p1",
		Text:     "x",
		Channels: model.ChannelList{model.ChannelTelegram},
		Status:   model.PostStatusPendingReview,
	})

	post, err := s.RejectPost(context.Background(), "p1", "bob", "typo in the headline")
	if err != nil {
		t.Fatalf("RejectPost: %v", err)
	}
	if post.Status != model.PostStatusDraft {
		t.Fatalf("expected draft, got %q", post.Status)
	}
	if post.RejectReason != "typo in the headline" {
		t.Fatalf("expected reject reason to be recorded, got %q", post.RejectReason)
	}
	if post.StatusChangedBy != "bob" {
		t.Fatalf("expected StatusChangedBy to be recorded, got %q", post.StatusChangedBy)
	}
}

func TestCancelScheduledPost_AllowsNonTerminalStatuses(t *testing.T) {
	s, repo, _ := newTestService(false)
	for _, status := range []model.PostStatus{
		model.PostStatusDraft, model.PostStatusPendingReview, model.PostStatusScheduled, model.PostStatusActive,
	} {
		id := string(status)
		repo.seed(model.Post{ID: id, Text: "x", Status: status})
		if err := s.CancelScheduledPost(context.Background(), id, "alice"); err != nil {
			t.Fatalf("expected cancel to succeed from status %q, got error: %v", status, err)
		}
	}
}

func TestCancelScheduledPost_RejectsTerminalStatuses(t *testing.T) {
	s, repo, _ := newTestService(false)
	for _, status := range []model.PostStatus{
		model.PostStatusPublished, model.PostStatusFailed, model.PostStatusCancelled,
	} {
		id := string(status)
		repo.seed(model.Post{ID: id, Text: "x", Status: status})
		if err := s.CancelScheduledPost(context.Background(), id, "alice"); err == nil {
			t.Fatalf("expected cancel to fail from terminal status %q", status)
		}
	}
}

// PublishDuePosts must never auto-publish anything but Scheduled/Active
// posts, even if the repository query returned something else (fakeRepo's
// DueForPublish/ActiveRecurring don't filter by status - see their comment -
// so this test exercises contentService's own guard, independent of SQL).
func TestPublishDuePosts_IgnoresPostsNotExactlyScheduledOrActive(t *testing.T) {
	s, repo, pub := newTestService(false)
	repo.seed(model.Post{
		ID: "draft-with-past-schedule", Channels: model.ChannelList{model.ChannelTelegram},
		Status: model.PostStatusDraft, ScheduledAt: time.Now().Add(-time.Hour),
	})
	repo.seed(model.Post{
		ID: "cancelled-recurring", Channels: model.ChannelList{model.ChannelTelegram},
		Status: model.PostStatusCancelled, CronExpr: "@every 1s",
	})

	if err := s.PublishDuePosts(context.Background()); err != nil {
		t.Fatalf("PublishDuePosts: %v", err)
	}
	if pub.calls != 0 {
		t.Fatalf("expected no publish calls for non-Scheduled/Active posts, got %d", pub.calls)
	}
	if len(repo.updated) != 0 {
		t.Fatalf("expected neither post to be touched, got %d updates", len(repo.updated))
	}
}

// The mirror image: a genuinely Scheduled post with a past ScheduledAt must
// still go out via the automated path.
func TestPublishDuePosts_PublishesScheduledPost(t *testing.T) {
	s, repo, pub := newTestService(false)
	repo.seed(model.Post{
		ID: "p1", Channels: model.ChannelList{model.ChannelTelegram},
		Status: model.PostStatusScheduled, ScheduledAt: time.Now().Add(-time.Hour),
	})

	if err := s.PublishDuePosts(context.Background()); err != nil {
		t.Fatalf("PublishDuePosts: %v", err)
	}
	if pub.calls != 1 {
		t.Fatalf("expected the due scheduled post to be published, got %d calls", pub.calls)
	}
}
