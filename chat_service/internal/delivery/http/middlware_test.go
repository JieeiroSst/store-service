package http

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/chat-service/config"
)

func TestAccessApi(t *testing.T) {
	h := &Http{Config: &config.Config{Secret: config.SecretConfig{AuthorizeKey: "secret"}}}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	cases := map[string]int{
		base64.StdEncoding.EncodeToString([]byte("secret")): http.StatusOK,
		base64.StdEncoding.EncodeToString([]byte("SECRET")): http.StatusForbidden,
		"": http.StatusForbidden,
		"not-base64": http.StatusForbidden,
	}
	for header, want := range cases {
		r := httptest.NewRequest(http.MethodGet, "/message/1", nil)
		if header != "" {
			r.Header.Set("Authorization", header)
		}
		w := httptest.NewRecorder()
		h.AccessApi(next).ServeHTTP(w, r)
		if w.Code != want {
			t.Errorf("header %q: got %d want %d", header, w.Code, want)
		}
	}
}
