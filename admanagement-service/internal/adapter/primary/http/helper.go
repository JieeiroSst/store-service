package http

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/gorilla/mux"
)

func parseID(r *http.Request) (uint, error) {
	return parsePathUint(r, "id")
}

func parsePathUint(r *http.Request, name string) (uint, error) {
	id, err := strconv.ParseUint(mux.Vars(r)[name], 10, 64)
	if err != nil {
		return 0, common.ErrInvalidInput
	}
	return uint(id), nil
}

func parseDate(raw string, fallback time.Time) time.Time {
	if raw == "" {
		return fallback
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return fallback
	}
	return t
}

func parseDateRange(r *http.Request) (from, to time.Time) {
	to = parseDate(r.URL.Query().Get("to"), time.Now())
	from = parseDate(r.URL.Query().Get("from"), to.AddDate(0, 0, -30))
	return from, to
}

// clientIP prefers the X-Forwarded-For header (set by upstream proxies/load
// balancers) and falls back to the raw connection address.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
