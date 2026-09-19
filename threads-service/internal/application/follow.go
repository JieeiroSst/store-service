package application

import (
	"context"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/google/uuid"
)

type followService struct {
	follows port.FollowRepository
	users   port.UserClient
}

func NewFollowService(follows port.FollowRepository, users port.UserClient) port.FollowUsecase {
	return &followService{follows: follows, users: users}
}

func (s *followService) Follow(ctx context.Context, followerID, followedID string) error {
	if followerID == followedID {
		return port.ErrCannotFollowSelf
	}
	exists, err := s.follows.Exists(ctx, followerID, followedID)
	if err != nil {
		return err
	}
	if exists {
		return port.ErrAlreadyFollowing
	}
	return s.follows.Create(ctx, &model.Follow{ID: uuid.NewString(), FollowerID: followerID, FollowedID: followedID})
}

func (s *followService) Unfollow(ctx context.Context, followerID, followedID string) error {
	return s.follows.Delete(ctx, followerID, followedID)
}

func (s *followService) ListFollowers(ctx context.Context, userID string, cursor string, limit int) ([]model.Author, string, error) {
	follows, next, err := s.follows.ListFollowers(ctx, userID, cursor, limit)
	if err != nil {
		return nil, "", err
	}
	ids := make([]string, len(follows))
	for i, f := range follows {
		ids[i] = f.FollowerID
	}
	authors, err := s.resolveAuthors(ctx, ids)
	return authors, next, err
}

func (s *followService) ListFollowing(ctx context.Context, userID string, cursor string, limit int) ([]model.Author, string, error) {
	follows, next, err := s.follows.ListFollowing(ctx, userID, cursor, limit)
	if err != nil {
		return nil, "", err
	}
	ids := make([]string, len(follows))
	for i, f := range follows {
		ids[i] = f.FollowedID
	}
	authors, err := s.resolveAuthors(ctx, ids)
	return authors, next, err
}

func (s *followService) resolveAuthors(ctx context.Context, ids []string) ([]model.Author, error) {
	if len(ids) == 0 {
		return []model.Author{}, nil
	}
	authors, err := s.users.GetUsers(ctx, ids)
	if err != nil {
		return nil, err
	}
	// Preserve list order (and skip IDs user_service couldn't resolve)
	// rather than returning the map in arbitrary order.
	result := make([]model.Author, 0, len(ids))
	for _, id := range ids {
		if a, ok := authors[id]; ok {
			result = append(result, *a)
		}
	}
	return result, nil
}
