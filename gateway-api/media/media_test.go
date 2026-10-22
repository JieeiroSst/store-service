package media

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	var gotPath, gotAuth, gotRange string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth, gotRange = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Range")
		w.Header().Set("Set-Cookie", "a=b")
		w.WriteHeader(http.StatusPartialContent)
	}))
	defer upstream.Close()
	h := Handler("/media/", func(service string) string {
		if service == "video-service" {
			return upstream.URL
		}
		return ""
	})

	cases := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/media/video-service/api/videos/abc-123/stream", http.StatusPartialContent},
		{http.MethodHead, "/media/video-service/api/videos/abc/hls/master.m3u8", http.StatusPartialContent},
		{http.MethodGet, "/media/video-service/api/videos/abc/hls/v1/seg_00012.ts", http.StatusPartialContent},
		{http.MethodGet, "/media/video-service/api/videos/abc", http.StatusNotFound},
		{http.MethodGet, "/media/video-service/api/videos/abc/hls/../../secret", http.StatusNotFound},
		{http.MethodGet, "/media/user-service/api/users", http.StatusNotFound},
		{http.MethodDelete, "/media/video-service/api/videos/abc/stream", http.StatusMethodNotAllowed},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, nil)
		req.Header.Set("Authorization", "Bearer secret")
		req.Header.Set("Range", "bytes=0-99")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Fatalf("%s %s: got %d, want %d", c.method, c.path, rec.Code, c.want)
		}
		if c.want == http.StatusPartialContent {
			if gotAuth != "" || gotRange != "bytes=0-99" || rec.Header().Get("Set-Cookie") != "" {
				t.Fatalf("%s: auth=%q range=%q cookie=%q", c.path, gotAuth, gotRange, rec.Header().Get("Set-Cookie"))
			}
		}
	}
	if gotPath != "/api/videos/abc/hls/v1/seg_00012.ts" {
		t.Fatalf("upstream path %q", gotPath)
	}
}
