package userclient

import (
	"context"
	"sync"

	"github.com/JIeeiroSst/threads-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

// httpUserClient resolves post authors through user-service's FindUser.
type httpUserClient struct {
	users *userservice.Client
}

func NewUserClient(users *userservice.Client) port.UserClient {
	return &httpUserClient{users: users}
}

func (c *httpUserClient) GetUser(ctx context.Context, userID string) (*model.Author, error) {
	u, err := c.users.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &model.Author{ID: u.ID, Username: u.Username}, nil
}

func (c *httpUserClient) GetUsers(ctx context.Context, userIDs []string) (map[string]*model.Author, error) {
	result := make(map[string]*model.Author, len(userIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range userIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			author, err := c.GetUser(ctx, id)
			if err != nil {
				return
			}
			mu.Lock()
			result[id] = author
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	return result, nil
}
