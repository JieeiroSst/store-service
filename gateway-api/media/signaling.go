package media

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/JIeeiroSst/gateway-api/auth"
)

const maxRoomID = 64

func Signaling(service string, base Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		room := r.URL.Query().Get("room_id")
		if !validID(room) || !validID(user.ID) {
			http.Error(w, "invalid room_id", http.StatusBadRequest)
			return
		}
		if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)
			return
		}
		target, err := url.Parse(base(service))
		if err != nil || target.Host == "" {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(target)
				pr.Out.URL.Path = "/ws"
				pr.Out.URL.RawPath = ""
				pr.Out.URL.RawQuery = url.Values{"user_id": {user.ID}, "room_id": {room}}.Encode()
				pr.Out.Header.Del("Authorization")
				pr.Out.Header.Del("Cookie")
				pr.Out.Header.Del("Origin")
			},
		}
		proxy.ServeHTTP(w, r)
	})
}

func validID(id string) bool {
	if id == "" || len(id) > maxRoomID || !utf8.ValidString(id) {
		return false
	}
	for _, c := range id {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

func Socket(service, path string, base Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Token == "" {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)
			return
		}
		target, err := url.Parse(base(service))
		if err != nil || target.Host == "" {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(target)
				pr.Out.URL.Path = path
				pr.Out.URL.RawPath = ""
				pr.Out.URL.RawQuery = ""
				pr.Out.Header.Del("Cookie")
				pr.Out.Header.Del("Origin")
				pr.Out.Header.Set("Authorization", "Bearer "+user.Token)
			},
		}
		proxy.ServeHTTP(w, r)
	})
}
