package application

import (
	"context"
	"sync"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

const identityTTL = 30 * time.Second

type identityEntry struct {
	userID  string
	expires time.Time
}

// identityService authenticates bearer tokens against user-service, keeping
// successful lookups for a short while so a busy client does not turn every
// request into a user-service call. Failures are never cached: a logged-out
// token must fail on the next request that user-service sees it.
type identityService struct {
	users port.UserDirectory
	now   func() time.Time

	mu    sync.Mutex
	cache map[string]identityEntry
}

func NewIdentityService(users port.UserDirectory) *identityService {
	return &identityService{users: users, now: time.Now, cache: map[string]identityEntry{}}
}

func (s *identityService) Authenticate(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", port.ErrUnauthenticated
	}
	s.mu.Lock()
	e, ok := s.cache[token]
	s.mu.Unlock()
	if ok && s.now().Before(e.expires) {
		return e.userID, nil
	}

	userID, err := s.users.Validate(ctx, token)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	if len(s.cache) > 10_000 {
		s.cache = map[string]identityEntry{}
	}
	s.cache[token] = identityEntry{userID: userID, expires: s.now().Add(identityTTL)}
	s.mu.Unlock()
	return userID, nil
}
