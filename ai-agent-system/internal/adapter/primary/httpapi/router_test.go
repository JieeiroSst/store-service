package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

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

type fakeHistory struct {
	entries []model.HistoryEntry
	err     error
	user    string
	limit   int
}

func (f *fakeHistory) List(_ context.Context, userID string, limit int) ([]model.HistoryEntry, error) {
	f.user, f.limit = userID, limit
	return f.entries, f.err
}

func serve(t *testing.T, d Dependencies, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	NewRouter(d).ServeHTTP(rec, req)
	return rec
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorBody {
	t.Helper()
	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body.Error
}

func TestHistoryRequiresUserID(t *testing.T) {
	rec := serve(t, Dependencies{History: &fakeHistory{}}, httptest.NewRequest(http.MethodGet, "/v1/chat/history", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	if e := decodeError(t, rec); e.Type != "bad_request" || e.Message != "X-User-Id header is required" {
		t.Fatalf("error = %+v", e)
	}
}

func TestHistoryRejectsBadLimit(t *testing.T) {
	for _, limit := range []string{"0", "-3", "abc"} {
		req := httptest.NewRequest(http.MethodGet, "/v1/chat/history?limit="+limit, nil)
		req.Header.Set(userIDHeader, "alice")

		if rec := serve(t, Dependencies{History: &fakeHistory{}}, req); rec.Code != http.StatusBadRequest {
			t.Fatalf("limit=%s: status = %d", limit, rec.Code)
		}
	}
}

func TestHistoryReturnsMessages(t *testing.T) {
	created := time.Date(2026, 10, 21, 8, 30, 0, 123e6, time.FixedZone("ICT", 7*3600))
	hist := &fakeHistory{entries: []model.HistoryEntry{
		{Role: model.RoleAssistant, Content: "hello", Source: "ollama", CreatedAt: created},
		{Role: model.RoleUser, Content: "hi", CreatedAt: created},
	}}
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/history?limit=20", nil)
	req.Header.Set(userIDHeader, "alice")

	rec := serve(t, Dependencies{History: hist}, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	if hist.user != "alice" || hist.limit != 20 {
		t.Fatalf("service got (%q, %d)", hist.user, hist.limit)
	}
	want := `{"messages":[{"role":"assistant","content":"hello","source":"ollama","created_at":"2026-10-21T01:30:00.123Z"},{"role":"user","content":"hi","created_at":"2026-10-21T01:30:00.123Z"}]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Fatalf("body = %s\nwant  %s", got, want)
	}
}

func TestHistoryMapsServiceError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/history", nil)
	req.Header.Set(userIDHeader, "alice")

	rec := serve(t, Dependencies{History: &fakeHistory{err: errors.New("db down")}}, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if e := decodeError(t, rec); e.Type != "internal" || strings.Contains(e.Message, "db down") {
		t.Fatalf("error leaks internals: %+v", e)
	}
}

func TestHealthEndpoints(t *testing.T) {
	ok := Dependencies{Ready: func(context.Context) error { return nil }}
	down := Dependencies{Ready: func(context.Context) error { return errors.New("no db") }}

	if rec := serve(t, down, httptest.NewRequest(http.MethodGet, "/healthz", nil)); rec.Code != http.StatusOK {
		t.Fatalf("healthz = %d", rec.Code)
	}
	if rec := serve(t, ok, httptest.NewRequest(http.MethodGet, "/readyz", nil)); rec.Code != http.StatusOK {
		t.Fatalf("readyz ok = %d", rec.Code)
	}
	if rec := serve(t, down, httptest.NewRequest(http.MethodGet, "/readyz", nil)); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz down = %d", rec.Code)
	}
}

func dialWS(t *testing.T, chat *fakeChat) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(NewRouter(Dependencies{Chat: chat}))
	t.Cleanup(srv.Close)

	header := http.Header{}
	header.Set(userIDHeader, "alice")
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/v1/chat/ws", header)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	return conn
}

func readFrame(t *testing.T, conn *websocket.Conn) wsMessage {
	t.Helper()
	var msg wsMessage
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read: %v", err)
	}
	return msg
}

func TestWebSocketStreamsTurn(t *testing.T) {
	chat := &fakeChat{
		deltas: []string{"he", "llo"},
		answer: model.Answer{Completion: model.Completion{Text: "hello"}, Source: model.SourceOllama},
	}
	conn := dialWS(t, chat)

	if err := conn.WriteJSON(wsRequest{Question: "hi"}); err != nil {
		t.Fatalf("write: %v", err)
	}

	want := []wsMessage{
		{Type: "delta", Text: "he"},
		{Type: "delta", Text: "llo"},
		{Type: "done", Answer: "hello", Source: "ollama"},
	}
	for _, w := range want {
		if got := readFrame(t, conn); got != w {
			t.Fatalf("frame = %+v, want %+v", got, w)
		}
	}
	if chat.got.UserID != "alice" || chat.got.LastUserQuestion() != "hi" {
		t.Fatalf("service got %+v", chat.got)
	}
}

func TestWebSocketReportsErrorsAndStaysOpen(t *testing.T) {
	chat := &fakeChat{err: model.NewError(model.KindUpstreamUnavailable, "down", nil)}
	conn := dialWS(t, chat)

	_ = conn.WriteJSON(wsRequest{Question: ""})
	if got := readFrame(t, conn); got.Type != "error" || got.Message != "question must not be empty" {
		t.Fatalf("frame = %+v", got)
	}

	_ = conn.WriteJSON(wsRequest{Question: "hi"})
	if got := readFrame(t, conn); got.Type != "error" || got.Message != "the AI service is temporarily unavailable" {
		t.Fatalf("frame = %+v", got)
	}
}
