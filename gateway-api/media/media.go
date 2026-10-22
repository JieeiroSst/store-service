package media

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

type Resolver func(service string) string

var rules = map[string]*regexp.Regexp{
	"video-service": regexp.MustCompile(`^/api/videos/[A-Za-z0-9_-]{1,64}/(stream|thumbnail|hls/(master\.m3u8|v\d{1,2}/(index\.m3u8|seg_\d{4,6}\.ts)))$`),
}

var forwarded = []string{"Range", "If-Range", "If-None-Match", "If-Modified-Since", "Accept", "Accept-Encoding", "User-Agent"}

func Handler(prefix string, base Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		service, path, ok := strings.Cut(strings.TrimPrefix(r.URL.Path, prefix), "/")
		path = "/" + path
		rule, known := rules[service]
		if !ok || !known || strings.Contains(path, "..") || !rule.MatchString(path) {
			http.NotFound(w, r)
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
				pr.Out.URL.Path = "/" + strings.TrimPrefix(target.JoinPath(path).Path, "/")
				pr.Out.URL.RawPath = ""
				pr.Out.URL.RawQuery = ""
				for k := range pr.Out.Header {
					if !slices.Contains(forwarded, http.CanonicalHeaderKey(k)) {
						pr.Out.Header.Del(k)
					}
				}
			},
			ModifyResponse: func(res *http.Response) error {
				res.Header.Del("Set-Cookie")
				for k := range res.Header {
					if strings.HasPrefix(http.CanonicalHeaderKey(k), "Access-Control-") {
						res.Header.Del(k)
					}
				}
				res.Header.Set("Access-Control-Expose-Headers", "Content-Range, Content-Length, Accept-Ranges, ETag")
				return nil
			},
		}
		proxy.ServeHTTP(w, r)
	})
}
