package ollamaapi

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aiagentsystem/internal/domain/model"
)

type fakeChat struct {
	deltas []string
	answer model.Answer
	err    error
	got    model.Question
}

func (f *fakeChat) Answer(_ context.Context, q model.Question) (model.Answer, error) {
	f.got = q
	return f.answer, f.err
}

func (f *fakeChat) AnswerStream(_ context.Context, q model.Question, onDelta func(string)) (model.Answer, error) {
	f.got = q
	for _, d := range f.deltas {
		onDelta(d)
	}
	return f.answer, f.err
}

func post(chat *fakeChat, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set(userIDHeader, "alice")
	rec := httptest.NewRecorder()
	NewRouter(NewHandlers(chat, "custom-model")).ServeHTTP(rec, req)
	return rec
}

func ndjsonLines(t *testing.T, body string) []map[string]any {
	t.Helper()
	var lines []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		var line map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatalf("bad ndjson line %q: %v", scanner.Text(), err)
		}
		lines = append(lines, line)
	}
	return lines
}

func TestChatNonStreaming(t *testing.T) {
	chat := &fakeChat{answer: model.Answer{Completion: model.Completion{Text: "hello", InputTokens: 2, OutputTokens: 3}}}

	rec := post(chat, "/api/chat", `{"model":"anything","messages":[{"role":"system","content":"be nice"},{"role":"user","content":"hi"}],"stream":false}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp ChatResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Model != "custom-model" || !resp.Done || resp.Message.Content != "hello" || resp.EvalCount != 3 || resp.PromptEvalCount != 2 {
		t.Fatalf("resp = %+v", resp)
	}
	if chat.got.UserID != "alice" || len(chat.got.Messages) != 2 || chat.got.Messages[0].Role != "system" {
		t.Fatalf("service got %+v", chat.got)
	}
}

func TestChatStreamsNDJSONByDefault(t *testing.T) {
	chat := &fakeChat{deltas: []string{"he", "llo"}, answer: model.Answer{Completion: model.Completion{Text: "hello", OutputTokens: 2}}}

	rec := post(chat, "/api/chat", `{"messages":[{"role":"user","content":"hi"}]}`)

	if ct := rec.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Fatalf("content-type = %q", ct)
	}
	lines := ndjsonLines(t, rec.Body.String())
	if len(lines) != 3 {
		t.Fatalf("lines = %v", lines)
	}
	if lines[0]["message"].(map[string]any)["content"] != "he" || lines[0]["done"] != false {
		t.Fatalf("first line = %v", lines[0])
	}
	last := lines[2]
	if last["done"] != true || last["message"].(map[string]any)["content"] != "" || last["eval_count"] != float64(2) {
		t.Fatalf("final line = %v", last)
	}
}

func TestStreamErrorIsFinalLine(t *testing.T) {
	chat := &fakeChat{deltas: []string{"par"}, err: model.NewError(model.KindRateLimited, "slow", nil)}

	rec := post(chat, "/api/generate", `{"prompt":"hi"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	lines := ndjsonLines(t, rec.Body.String())
	if len(lines) != 2 || lines[0]["response"] != "par" || lines[1]["error"] != "rate limit exceeded, please try again" {
		t.Fatalf("lines = %v", lines)
	}
}

func TestGenerateNonStreaming(t *testing.T) {
	chat := &fakeChat{answer: model.Answer{Completion: model.Completion{Text: "blue"}}}

	rec := post(chat, "/api/generate", `{"prompt":"why is the sky blue?","stream":false}`)

	var resp GenerateResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if rec.Code != http.StatusOK || resp.Response != "blue" || !resp.Done {
		t.Fatalf("status = %d, resp = %+v", rec.Code, resp)
	}
	if chat.got.LastUserQuestion() != "why is the sky blue?" {
		t.Fatalf("service got %+v", chat.got)
	}
}

func TestErrorStatusMapping(t *testing.T) {
	cases := []struct {
		kind   model.ErrorKind
		status int
	}{
		{model.KindRateLimited, http.StatusTooManyRequests},
		{model.KindUpstreamUnavailable, http.StatusServiceUnavailable},
		{model.KindBadRequest, http.StatusBadRequest},
		{model.KindInternal, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		chat := &fakeChat{err: model.NewError(tc.kind, "x", nil)}
		if rec := post(chat, "/api/chat", `{"messages":[{"role":"user","content":"hi"}],"stream":false}`); rec.Code != tc.status {
			t.Fatalf("kind %v: status = %d, want %d", tc.kind, rec.Code, tc.status)
		}
	}
}

func TestRequestValidation(t *testing.T) {
	cases := map[string]string{
		"/api/generate": `{"prompt":""}`,
		"/api/chat":     `{"messages":[]}`,
		"/api/show":     `not json`,
	}
	for path, body := range cases {
		if rec := post(&fakeChat{}, path, body); rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d", path, rec.Code)
		}
	}
}

func TestTagsAndShow(t *testing.T) {
	rec := httptest.NewRecorder()
	NewRouter(NewHandlers(&fakeChat{}, "custom-model")).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tags", nil))
	var tags TagsResponse
	_ = json.NewDecoder(rec.Body).Decode(&tags)
	if len(tags.Models) != 1 || tags.Models[0].Name != "custom-model" || tags.Models[0].Digest == "" {
		t.Fatalf("tags = %+v", tags)
	}

	if rec := post(&fakeChat{}, "/api/show", `{"model":"custom-model"}`); rec.Code != http.StatusOK {
		t.Fatalf("show known model: %d", rec.Code)
	}
	if rec := post(&fakeChat{}, "/api/show", `{"model":"llama3"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("show unknown model: %d", rec.Code)
	}
}
