package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"aiagentsystem/internal/domain/model"
)

func newTestClient(url string) *Client {
	return New(Config{BaseURL: url + "/", Model: "custom-model", RetryBaseDelay: time.Millisecond})
}

func hi(userID string) model.Question {
	return model.Question{UserID: userID, Messages: []model.Message{{Role: model.RoleUser, Content: "hi"}}}
}

func TestCompleteSendsChatRequest(t *testing.T) {
	var got chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		fmt.Fprint(w, `{"message":{"role":"assistant","content":"hello"},"done":true,"done_reason":"stop","prompt_eval_count":4,"eval_count":7}`)
	}))
	defer srv.Close()

	completion, err := newTestClient(srv.URL).Complete(context.Background(), hi("alice"))
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	want := model.Completion{Text: "hello", StopReason: "stop", InputTokens: 4, OutputTokens: 7}
	if completion != want {
		t.Fatalf("completion = %+v, want %+v", completion, want)
	}
	if got.Model != "custom-model" || got.Stream || got.UserID != "alice" || len(got.Messages) != 1 || got.Messages[0].Content != "hi" {
		t.Fatalf("request = %+v", got)
	}
}

func TestCompleteMapsStatusCodes(t *testing.T) {
	cases := []struct {
		status   int
		kind     model.ErrorKind
		attempts int32
	}{
		{http.StatusNotFound, model.KindBadRequest, 1},
		{http.StatusBadRequest, model.KindBadRequest, 1},
		{http.StatusTooManyRequests, model.KindRateLimited, 3},
		{http.StatusBadGateway, model.KindUpstreamUnavailable, 3},
		{http.StatusForbidden, model.KindInternal, 1},
	}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			var attempts atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				http.Error(w, "nope", tc.status)
			}))
			defer srv.Close()

			_, err := newTestClient(srv.URL).Complete(context.Background(), hi(""))

			if got := model.AsError(err).Kind; got != tc.kind {
				t.Fatalf("kind = %v, want %v (err %v)", got, tc.kind, err)
			}
			if got := attempts.Load(); got != tc.attempts {
				t.Fatalf("attempts = %d, want %d", got, tc.attempts)
			}
		})
	}
}

func TestCompleteRetriesTransientFailure(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprint(w, `{"message":{"content":"ok"},"done":true}`)
	}))
	defer srv.Close()

	completion, err := newTestClient(srv.URL).Complete(context.Background(), hi(""))

	if err != nil || completion.Text != "ok" {
		t.Fatalf("completion = %+v, err = %v", completion, err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}
}

func TestCompleteUnreachableBackend(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	_, err := newTestClient(url).Complete(context.Background(), hi(""))

	if got := model.AsError(err).Kind; got != model.KindUpstreamUnavailable {
		t.Fatalf("kind = %v, want upstream unavailable", got)
	}
}

func TestStreamYieldsDeltasAndFinalUsage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if !req.Stream {
			t.Error("stream flag not set")
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		fmt.Fprintln(w, `{"message":{"content":"he"},"done":false}`)
		fmt.Fprintln(w, ``)
		fmt.Fprintln(w, `{"message":{"content":""},"done":false}`)
		fmt.Fprintln(w, `{"message":{"content":"llo"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"content":""},"done":true,"done_reason":"stop","prompt_eval_count":2,"eval_count":3}`)
	}))
	defer srv.Close()

	stream, err := newTestClient(srv.URL).Stream(context.Background(), hi(""))
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	var deltas []string
	for stream.Next() {
		deltas = append(deltas, stream.Current())
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream err: %v", err)
	}
	if len(deltas) != 2 || deltas[0] != "he" || deltas[1] != "llo" {
		t.Fatalf("deltas = %v", deltas)
	}
	want := model.Completion{Text: "hello", StopReason: "stop", InputTokens: 2, OutputTokens: 3}
	if got := stream.Result(); got != want {
		t.Fatalf("result = %+v, want %+v", got, want)
	}
}

func TestStreamMalformedChunk(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `not json`)
	}))
	defer srv.Close()

	stream, err := newTestClient(srv.URL).Stream(context.Background(), hi(""))
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	if stream.Next() {
		t.Fatal("Next() = true on malformed chunk")
	}
	if got := model.AsError(stream.Err()).Kind; got != model.KindInternal {
		t.Fatalf("kind = %v, want internal", got)
	}
}
