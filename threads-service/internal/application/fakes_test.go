package application

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

type fakePostRepository struct {
	posts map[string]*model.Post
}

func newFakePostRepository() *fakePostRepository {
	return &fakePostRepository{posts: make(map[string]*model.Post)}
}

func (f *fakePostRepository) Create(ctx context.Context, post *model.Post) error {
	f.posts[post.ID] = post
	return nil
}

func (f *fakePostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	post, ok := f.posts[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	got := *post
	return &got, nil
}

// ListFeed and every other fake List* method below ignore cursor/limit
// and just return every matching row with no next page - these tests
// exercise business logic (ownership, dedup, filtering), not pagination
// slicing, which is covered at the repository level instead.
func (f *fakePostRepository) ListFeed(ctx context.Context, authorID string, cursor string, limit int) ([]model.Post, string, error) {
	var result []model.Post
	for _, post := range f.posts {
		if authorID == "" || post.UserID == authorID {
			result = append(result, *post)
		}
	}
	return result, "", nil
}

func (f *fakePostRepository) ListByAuthors(ctx context.Context, authorIDs []string, cursor string, limit int) ([]model.Post, string, error) {
	authorSet := make(map[string]bool, len(authorIDs))
	for _, id := range authorIDs {
		authorSet[id] = true
	}
	var result []model.Post
	for _, post := range f.posts {
		if authorSet[post.UserID] {
			result = append(result, *post)
		}
	}
	return result, "", nil
}

func (f *fakePostRepository) Delete(ctx context.Context, id string) error {
	delete(f.posts, id)
	return nil
}

func (f *fakePostRepository) IncrementLikeCount(ctx context.Context, id string, delta int) error {
	if post, ok := f.posts[id]; ok {
		post.LikeCount += delta
	}
	return nil
}

func (f *fakePostRepository) IncrementCommentCount(ctx context.Context, id string, delta int) error {
	if post, ok := f.posts[id]; ok {
		post.CommentCount += delta
	}
	return nil
}

func (f *fakePostRepository) IncrementRepostCount(ctx context.Context, id string, delta int) error {
	if post, ok := f.posts[id]; ok {
		post.RepostCount += delta
	}
	return nil
}

func (f *fakePostRepository) FindRepostBy(ctx context.Context, userID, originalPostID string) (*model.Post, error) {
	for _, post := range f.posts {
		if post.UserID == userID && post.RepostOfID != nil && *post.RepostOfID == originalPostID {
			got := *post
			return &got, nil
		}
	}
	return nil, port.ErrNotFound
}

type fakeCommentRepository struct {
	comments map[string]*model.Comment
}

func newFakeCommentRepository() *fakeCommentRepository {
	return &fakeCommentRepository{comments: make(map[string]*model.Comment)}
}

func (f *fakeCommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	f.comments[comment.ID] = comment
	return nil
}

func (f *fakeCommentRepository) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	c, ok := f.comments[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	got := *c
	return &got, nil
}

func (f *fakeCommentRepository) ListByPost(ctx context.Context, postID string, cursor string, limit int) ([]model.Comment, string, error) {
	var result []model.Comment
	for _, c := range f.comments {
		if c.PostID == postID {
			result = append(result, *c)
		}
	}
	return result, "", nil
}

func (f *fakeCommentRepository) Delete(ctx context.Context, id string) error {
	delete(f.comments, id)
	return nil
}

type fakeLikeRepository struct {
	likes map[string]model.Like
}

func newFakeLikeRepository() *fakeLikeRepository {
	return &fakeLikeRepository{likes: make(map[string]model.Like)}
}

func likeKey(userID string, postID, commentID *string) string {
	switch {
	case postID != nil:
		return "post:" + userID + ":" + *postID
	case commentID != nil:
		return "comment:" + userID + ":" + *commentID
	}
	return ""
}

func (f *fakeLikeRepository) Create(ctx context.Context, like *model.Like) error {
	f.likes[likeKey(like.UserID, like.PostID, like.CommentID)] = *like
	return nil
}

func (f *fakeLikeRepository) Delete(ctx context.Context, userID string, postID, commentID *string) (bool, error) {
	key := likeKey(userID, postID, commentID)
	if _, ok := f.likes[key]; !ok {
		return false, nil
	}
	delete(f.likes, key)
	return true, nil
}

func (f *fakeLikeRepository) Exists(ctx context.Context, userID string, postID, commentID *string) (bool, error) {
	_, ok := f.likes[likeKey(userID, postID, commentID)]
	return ok, nil
}

type fakeFollowRepository struct {
	follows map[string]model.Follow
}

func newFakeFollowRepository() *fakeFollowRepository {
	return &fakeFollowRepository{follows: make(map[string]model.Follow)}
}

func followKey(followerID, followedID string) string { return followerID + ":" + followedID }

func (f *fakeFollowRepository) Create(ctx context.Context, follow *model.Follow) error {
	f.follows[followKey(follow.FollowerID, follow.FollowedID)] = *follow
	return nil
}

func (f *fakeFollowRepository) Delete(ctx context.Context, followerID, followedID string) error {
	delete(f.follows, followKey(followerID, followedID))
	return nil
}

func (f *fakeFollowRepository) Exists(ctx context.Context, followerID, followedID string) (bool, error) {
	_, ok := f.follows[followKey(followerID, followedID)]
	return ok, nil
}

func (f *fakeFollowRepository) ListFollowers(ctx context.Context, userID string, cursor string, limit int) ([]model.Follow, string, error) {
	var result []model.Follow
	for _, flw := range f.follows {
		if flw.FollowedID == userID {
			result = append(result, flw)
		}
	}
	return result, "", nil
}

func (f *fakeFollowRepository) ListFollowing(ctx context.Context, userID string, cursor string, limit int) ([]model.Follow, string, error) {
	var result []model.Follow
	for _, flw := range f.follows {
		if flw.FollowerID == userID {
			result = append(result, flw)
		}
	}
	return result, "", nil
}

func (f *fakeFollowRepository) ListFollowedIDs(ctx context.Context, followerID string) ([]string, error) {
	var ids []string
	for _, flw := range f.follows {
		if flw.FollowerID == followerID {
			ids = append(ids, flw.FollowedID)
		}
	}
	return ids, nil
}

type fakeTagRepository struct {
	byName map[string]model.Tag
}

func newFakeTagRepository() *fakeTagRepository {
	return &fakeTagRepository{byName: make(map[string]model.Tag)}
}

func (f *fakeTagRepository) GetOrCreateByNames(ctx context.Context, names []string) ([]model.Tag, error) {
	tags := make([]model.Tag, 0, len(names))
	for _, name := range names {
		tag, ok := f.byName[name]
		if !ok {
			tag = model.Tag{ID: name, Name: name}
			f.byName[name] = tag
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (f *fakeTagRepository) AttachToPost(ctx context.Context, postID string, tagIDs []string) error {
	return nil
}

func (f *fakeTagRepository) ListNamesByPost(ctx context.Context, postID string) ([]string, error) {
	return nil, nil
}

type fakeBookmarkRepository struct {
	bookmarks map[string]model.Bookmark
}

func newFakeBookmarkRepository() *fakeBookmarkRepository {
	return &fakeBookmarkRepository{bookmarks: make(map[string]model.Bookmark)}
}

func bookmarkKey(userID, postID string) string { return userID + ":" + postID }

func (f *fakeBookmarkRepository) Create(ctx context.Context, bookmark *model.Bookmark) error {
	f.bookmarks[bookmarkKey(bookmark.UserID, bookmark.PostID)] = *bookmark
	return nil
}

func (f *fakeBookmarkRepository) Delete(ctx context.Context, userID, postID string) (bool, error) {
	key := bookmarkKey(userID, postID)
	if _, ok := f.bookmarks[key]; !ok {
		return false, nil
	}
	delete(f.bookmarks, key)
	return true, nil
}

func (f *fakeBookmarkRepository) Exists(ctx context.Context, userID, postID string) (bool, error) {
	_, ok := f.bookmarks[bookmarkKey(userID, postID)]
	return ok, nil
}

func (f *fakeBookmarkRepository) ListPostsByUser(ctx context.Context, userID string, cursor string, limit int) ([]model.Post, string, error) {
	return nil, "", nil
}

type fakeTrendingTagsStore struct {
	counts map[string]int64
}

func newFakeTrendingTagsStore() *fakeTrendingTagsStore {
	return &fakeTrendingTagsStore{counts: make(map[string]int64)}
}

func (f *fakeTrendingTagsStore) IncrementTags(ctx context.Context, names []string) error {
	for _, name := range names {
		f.counts[name]++
	}
	return nil
}

func (f *fakeTrendingTagsStore) TopTags(ctx context.Context, limit int) ([]model.TagCount, error) {
	tags := make([]model.TagCount, 0, len(f.counts))
	for name, count := range f.counts {
		tags = append(tags, model.TagCount{Name: name, Count: count})
	}
	return tags, nil
}

// newTestPostService wires a postService with fresh fakes for every
// dependency except posts/users, which most tests care about directly.
func newTestPostService(posts port.PostRepository, users port.UserClient) port.PostUsecase {
	return NewPostService(posts, newFakeTagRepository(), newFakeFollowRepository(), newFakeTrendingTagsStore(), users)
}

type fakeUserClient struct {
	users map[string]*model.Author
	err   error
}

func newFakeUserClient() *fakeUserClient {
	return &fakeUserClient{users: make(map[string]*model.Author)}
}

func (f *fakeUserClient) GetUser(ctx context.Context, userID string) (*model.Author, error) {
	if f.err != nil {
		return nil, f.err
	}
	author, ok := f.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return author, nil
}

func (f *fakeUserClient) GetUsers(ctx context.Context, userIDs []string) (map[string]*model.Author, error) {
	if f.err != nil {
		return nil, f.err
	}
	result := make(map[string]*model.Author)
	for _, id := range userIDs {
		if a, ok := f.users[id]; ok {
			result[id] = a
		}
	}
	return result, nil
}
