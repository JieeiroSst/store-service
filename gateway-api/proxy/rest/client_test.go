package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/gateway-api/auth"
)

func TestForwardsUserAccessToken(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()
	user := auth.WithUser(context.Background(), auth.User{ID: "7", Token: "session-token-from-login"})

	var out any
	if err := New("wishlists-service", srv.URL, time.Second).Get(user, "/api/v1/wishlists", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "Bearer session-token-from-login" || got.Get("X-User-Id") != "7" {
		t.Fatalf("user access token must be forwarded: %v", got)
	}

	if err := New("wishlists-service", srv.URL, time.Second).Get(context.Background(), "/api/v1/wishlist-stats/popular-products", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "" || got.Get("X-User-Id") != "" {
		t.Fatalf("anonymous calls must not carry credentials: %v", got)
	}

	if err := New("search-service", srv.URL, time.Second, Options{DropUserToken: true}).Get(user, "/", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "" {
		t.Fatalf("forward_user_token=false must drop the token: %v", got)
	}
}
