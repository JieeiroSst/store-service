package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/ollama-service/internal/userservice"
	"github.com/gin-gonic/gin"
)

// users is the only source of identity: accounts, sign-up and login live in
// user-service, and every bearer token is validated there.
var users = userservice.New(getEnv("USER_SERVICE_URL", "http://user-service:1235"), 3*time.Second)

const ctxKeyUserID = "userID"

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// bearerToken reads the token from the Authorization header, or from the
// "token" query parameter for WebSocket upgrades (browsers can't set headers
// on those).
func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return r.URL.Query().Get("token")
}

// authenticate resolves the request's token to a user-service user id and
// writes the HTTP error itself when it can't.
func authenticate(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := users.Authenticate(r.Context(), bearerToken(r))
	if errors.Is(err, userservice.ErrUpstream) {
		http.Error(w, `{"error":"authentication service unavailable"}`, http.StatusServiceUnavailable)
		return 0, false
	}
	if err != nil {
		http.Error(w, `{"error":"invalid or missing bearer token"}`, http.StatusUnauthorized)
		return 0, false
	}
	userID, err := strconv.ParseInt(id.UserID, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id in session"}`, http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

func requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := authenticate(c.Writer, c.Request)
		if !ok {
			c.Abort()
			return
		}
		c.Set(ctxKeyUserID, userID)
		c.Next()
	}
}

// callerID is the authenticated user; client-supplied ids are never trusted.
func callerID(c *gin.Context) (int64, error) {
	v, ok := c.Get(ctxKeyUserID)
	id, _ := v.(int64)
	if !ok || id == 0 {
		return 0, errors.New("unauthenticated")
	}
	return id, nil
}

// profile is the public part of a user-service account shown in chats.
type profile struct {
	Username string
	Email    string
}

// lookupProfiles fetches profiles for ids from user-service in parallel.
// Users that can't be resolved are simply missing from the map.
func lookupProfiles(ctx context.Context, ids []int64) map[int64]profile {
	out := make(map[int64]profile, len(ids))
	var mu sync.Mutex
	var wg sync.WaitGroup
	seen := map[int64]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			u, err := users.GetUser(ctx, strconv.FormatInt(id, 10))
			if err != nil {
				return
			}
			mu.Lock()
			out[id] = profile{Username: u.Username, Email: u.Email}
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return out
}
