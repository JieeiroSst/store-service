package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

func TestCreatePostSetsAuthorFromUserClient(t *testing.T) {
	users := newFakeUserClient()
	users.users["user-1"] = &model.Author{ID: "user-1", Username: "alice"}
	u := newTestPostService(newFakePostRepository(), users)

	post, err := u.CreatePost(context.Background(), model.CreatePostInput{UserID: "user-1", Content: "hello"})
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if post.Author == nil || post.Author.Username != "alice" {
		t.Errorf("Author = %+v, want username alice", post.Author)
	}
}

func TestCreatePostDegradesGracefullyWhenUserServiceIsDown(t *testing.T) {
	users := newFakeUserClient()
	users.err = context.DeadlineExceeded
	u := newTestPostService(newFakePostRepository(), users)

	post, err := u.CreatePost(context.Background(), model.CreatePostInput{UserID: "user-1", Content: "hello"})
	if err != nil {
		t.Fatalf("CreatePost() error = %v, want the post to still be created", err)
	}
	if post.Author != nil {
		t.Errorf("Author = %+v, want nil when user_service is unreachable", post.Author)
	}
}

func TestDeletePostRejectsNonOwner(t *testing.T) {
	u := newTestPostService(newFakePostRepository(), newFakeUserClient())
	post, err := u.CreatePost(context.Background(), model.CreatePostInput{UserID: "owner-1", Content: "hello"})
	if err != nil {
		t.Fatalf("setup CreatePost() error = %v", err)
	}

	if err := u.DeletePost(context.Background(), post.ID, "someone-else"); err != port.ErrForbidden {
		t.Fatalf("DeletePost() error = %v, want ErrForbidden", err)
	}
}

func TestDeletePostAllowsOwner(t *testing.T) {
	u := newTestPostService(newFakePostRepository(), newFakeUserClient())
	post, err := u.CreatePost(context.Background(), model.CreatePostInput{UserID: "owner-1", Content: "hello"})
	if err != nil {
		t.Fatalf("setup CreatePost() error = %v", err)
	}

	if err := u.DeletePost(context.Background(), post.ID, "owner-1"); err != nil {
		t.Fatalf("DeletePost() error = %v", err)
	}
}

func TestRepostRejectsDuplicateAndIncrementsCount(t *testing.T) {
	posts := newFakePostRepository()
	u := newTestPostService(posts, newFakeUserClient())
	original, err := u.CreatePost(context.Background(), model.CreatePostInput{UserID: "owner-1", Content: "original"})
	if err != nil {
		t.Fatalf("setup CreatePost() error = %v", err)
	}

	if _, err := u.Repost(context.Background(), "user-2", original.ID); err != nil {
		t.Fatalf("first Repost() error = %v", err)
	}
	if _, err := u.Repost(context.Background(), "user-2", original.ID); err != port.ErrAlreadyReposted {
		t.Fatalf("second Repost() error = %v, want ErrAlreadyReposted", err)
	}
	if got := posts.posts[original.ID].RepostCount; got != 1 {
		t.Errorf("RepostCount = %d, want 1", got)
	}
}

func TestUnrepostDecrementsCountAndAllowsReposting(t *testing.T) {
	posts := newFakePostRepository()
	u := newTestPostService(posts, newFakeUserClient())
	original, err := u.CreatePost(context.Background(), model.CreatePostInput{UserID: "owner-1", Content: "original"})
	if err != nil {
		t.Fatalf("setup CreatePost() error = %v", err)
	}
	if _, err := u.Repost(context.Background(), "user-2", original.ID); err != nil {
		t.Fatalf("Repost() error = %v", err)
	}

	if err := u.Unrepost(context.Background(), "user-2", original.ID); err != nil {
		t.Fatalf("Unrepost() error = %v", err)
	}
	if got := posts.posts[original.ID].RepostCount; got != 0 {
		t.Errorf("RepostCount = %d, want 0", got)
	}
	if _, err := u.Repost(context.Background(), "user-2", original.ID); err != nil {
		t.Fatalf("re-Repost() error = %v", err)
	}
}

func TestUnrepostWithoutARepostReturnsNotFound(t *testing.T) {
	posts := newFakePostRepository()
	u := newTestPostService(posts, newFakeUserClient())
	original, err := u.CreatePost(context.Background(), model.CreatePostInput{UserID: "owner-1", Content: "original"})
	if err != nil {
		t.Fatalf("setup CreatePost() error = %v", err)
	}

	if err := u.Unrepost(context.Background(), "user-2", original.ID); err != port.ErrNotFound {
		t.Fatalf("Unrepost() error = %v, want ErrNotFound", err)
	}
}

func TestListHomeFeedOnlyReturnsPostsFromFollowedUsers(t *testing.T) {
	posts := newFakePostRepository()
	follows := newFakeFollowRepository()
	users := newFakeUserClient()
	u := NewPostService(posts, newFakeTagRepository(), follows, newFakeTrendingTagsStore(), users)

	if err := follows.Create(context.Background(), &model.Follow{ID: "f1", FollowerID: "me", FollowedID: "followed-1"}); err != nil {
		t.Fatalf("setup follow error = %v", err)
	}
	posts.posts["p1"] = &model.Post{ID: "p1", UserID: "followed-1", Content: "from someone I follow"}
	posts.posts["p2"] = &model.Post{ID: "p2", UserID: "stranger", Content: "from a stranger"}

	result, _, err := u.ListHomeFeed(context.Background(), "me", "", 0)
	if err != nil {
		t.Fatalf("ListHomeFeed() error = %v", err)
	}
	if len(result) != 1 || result[0].ID != "p1" {
		t.Fatalf("ListHomeFeed() = %+v, want only post p1", result)
	}
}

func TestListHomeFeedIsEmptyWhenNotFollowingAnyone(t *testing.T) {
	u := newTestPostService(newFakePostRepository(), newFakeUserClient())

	result, _, err := u.ListHomeFeed(context.Background(), "me", "", 0)
	if err != nil {
		t.Fatalf("ListHomeFeed() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("ListHomeFeed() = %+v, want empty", result)
	}
}
