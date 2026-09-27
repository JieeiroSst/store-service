package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/learning"
	"github.com/JIeeiroSst/tool-service/internal/ollama"
)

func newTestHandler(t *testing.T, token string, allow []string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mem, err := learning.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := New(ollama.NewClient("http://unused", "m", 0), mem, LearnConfig{}, allow)
	h.SetAPIToken(token)
	h.SetInfra(Infra{DataDir: t.TempDir()})
	r := gin.New()
	h.Register(r)
	return r
}

func post(r *gin.Engine, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSensitiveEndpointsNeedToken(t *testing.T) {
	r := newTestHandler(t, "", []string{"127.0.0.1"})
	for _, p := range []string{"/api/v1/qa/database", "/api/v1/qa/active", "/api/v1/qa/queue"} {
		if w := post(r, p, "", `{"connection":"x","broker":"x","base_url":"http://127.0.0.1","active":true}`); w.Code != http.StatusForbidden {
			t.Errorf("%s without API_TOKEN: %d %s", p, w.Code, w.Body)
		}
	}
}

func TestActiveTestingSafeguards(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer target.Close()

	r := newTestHandler(t, "tok", []string{"127.0.0.1"})
	if w := post(r, "/api/v1/qa/active", "tok", `{"base_url":"`+target.URL+`"}`); w.Code != http.StatusBadRequest {
		t.Errorf("without active=true: %d", w.Code)
	}

	r = newTestHandler(t, "tok", nil)
	if w := post(r, "/api/v1/qa/active", "tok", `{"base_url":"`+target.URL+`","active":true}`); w.Code != http.StatusForbidden {
		t.Errorf("without TARGET_ALLOWLIST: %d", w.Code)
	}

	r = newTestHandler(t, "tok", []string{"only-this.example"})
	if w := post(r, "/api/v1/qa/active", "tok", `{"base_url":"`+target.URL+`","active":true}`); w.Code != http.StatusBadRequest {
		t.Errorf("target outside allow-list: %d", w.Code)
	}

	if w := post(r, "/api/v1/qa/active", "nope", `{}`); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong token: %d", w.Code)
	}
}

func TestDatabaseUnknownConnection(t *testing.T) {
	r := newTestHandler(t, "tok", nil)
	w := post(r, "/api/v1/qa/database", "tok", `{"connection":"missing"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
}

func TestStressJobEndToEnd(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(5 * time.Millisecond) }))
	defer target.Close()
	r := newTestHandler(t, "tok", nil)

	w := post(r, "/api/v1/jobs/stress", "tok", `{"url":"`+target.URL+`","stages":[2,4],"stage_sec":5,"max_p95_ms":2000}`)
	if w.Code != http.StatusAccepted {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var started struct {
		JobID string `json:"job_id"`
	}
	json.Unmarshal(w.Body.Bytes(), &started)

	deadline := time.Now().Add(40 * time.Second)
	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+started.JobID, nil)
		req.Header.Set("Authorization", "Bearer tok")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		var j struct {
			Status string `json:"status"`
			Result struct {
				Report struct {
					MaxSustainable int `json:"max_sustainable_concurrency"`
				} `json:"report"`
			} `json:"result"`
		}
		json.Unmarshal(rec.Body.Bytes(), &j)
		if j.Status == "done" {
			if j.Result.Report.MaxSustainable != 4 {
				t.Fatalf("both stages should pass: %s", rec.Body)
			}
			return
		}
		if j.Status == "failed" {
			t.Fatalf("job failed: %s", rec.Body)
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatal("job did not finish")
}

func TestArtifactRejectsTraversal(t *testing.T) {
	r := newTestHandler(t, "tok", nil)
	for _, name := range []string{"..%2f..%2fetc%2fpasswd", "a.txt", "..png"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/artifacts/"+name, nil)
		req.Header.Set("Authorization", "Bearer tok")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Errorf("%s must not be served", name)
		}
	}
}
