package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/adapter/primary/ws"
	"github.com/JIeeiroSst/room-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/room-service/internal/application"
	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"github.com/gorilla/websocket"
)

// directory accepts password "pw" for anyone and uses "tok-<name>" as the token.
type directory struct{}

func (directory) Login(_ context.Context, username, password string) (model.Session, error) {
	if password != "pw" {
		return model.Session{}, port.ErrInvalidLogin
	}
	return model.Session{AccessToken: "tok-" + username}, nil
}

func (directory) Refresh(_ context.Context, refreshToken string) (model.Session, error) {
	if refreshToken != "refresh-ok" {
		return model.Session{}, port.ErrUnauthenticated
	}
	return model.Session{AccessToken: "tok-alice", RefreshToken: "refresh-ok"}, nil
}

func (directory) Validate(_ context.Context, token string) (model.User, error) {
	name, ok := strings.CutPrefix(token, "tok-")
	if !ok {
		return model.User{}, port.ErrUnauthenticated
	}
	return model.User{Username: name}, nil
}

type localPublisher struct{ hub *ws.Hub }

func (p localPublisher) Publish(_ context.Context, e model.Event) error {
	p.hub.Deliver(e)
	return nil
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	cfg := &config.Config{}
	cfg.Server.AllowedOrigins = []string{"*"}
	cfg.Chat.MaxMessageLen = 100
	cfg.Chat.HistoryLimit = 50

	store := memory.NewStore()
	hub := ws.NewHub()
	pub := localPublisher{hub}
	h := NewHandler(
		application.NewAuthService(directory{}),
		application.NewRoomService(store, pub),
		application.NewChatService(store, memory.Messages{Store: store}, pub, cfg),
		hub, cfg,
	)
	srv := httptest.NewServer(NewRouter(h, cfg))
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var r bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&r).Encode(body)
	}
	req, _ := http.NewRequest(method, srv.URL+path, &r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out bytes.Buffer
	_, _ = out.ReadFrom(res.Body)
	return res.StatusCode, out.Bytes()
}

func dial(t *testing.T, srv *httptest.Server, room uint, token string) (*websocket.Conn, *http.Response, error) {
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/rooms/" + strconv.Itoa(int(room)) + "/ws?token=" + token
	conn, res, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		t.Cleanup(func() { conn.Close() })
	}
	return conn, res, err
}

func next(t *testing.T, c *websocket.Conn) model.Event {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	var e model.Event
	if err := c.ReadJSON(&e); err != nil {
		t.Fatalf("read: %v", err)
	}
	return e
}

func TestAuthRequired(t *testing.T) {
	srv := newServer(t)
	if code, _ := call(t, srv, "GET", "/api/rooms", "", nil); code != 401 {
		t.Fatalf("no token: %d", code)
	}
	if code, _ := call(t, srv, "GET", "/api/rooms", "garbage", nil); code != 401 {
		t.Fatalf("bad token: %d", code)
	}
	if code, _ := call(t, srv, "POST", "/api/auth/login", "", map[string]string{"username": "a", "password": "x"}); code != 401 {
		t.Fatalf("bad login: %d", code)
	}
	if code, _ := call(t, srv, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": "nope"}); code != 401 {
		t.Fatalf("bad refresh: %d", code)
	}
	if code, body := call(t, srv, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": "refresh-ok"}); code != 200 || !strings.Contains(string(body), "tok-alice") {
		t.Fatalf("refresh: %d %s", code, body)
	}
	code, body := call(t, srv, "POST", "/api/auth/login", "", map[string]string{"username": "alice", "password": "pw"})
	if code != 200 || !strings.Contains(string(body), "tok-alice") {
		t.Fatalf("login: %d %s", code, body)
	}
}

func TestChatFlow(t *testing.T) {
	srv := newServer(t)

	code, body := call(t, srv, "POST", "/api/rooms", "tok-alice", map[string]string{"name": "general"})
	var room model.Room
	_ = json.Unmarshal(body, &room)
	if code != 201 || room.ID == 0 {
		t.Fatalf("create: %d %s", code, body)
	}
	roomPath := "/api/rooms/" + strconv.Itoa(int(room.ID))

	// bob is not invited yet: refused before the upgrade
	if _, res, err := dial(t, srv, room.ID, "tok-bob"); err == nil || res.StatusCode != 403 {
		t.Fatalf("non-member dial: err=%v res=%v", err, res)
	}
	if code, _ := call(t, srv, "GET", roomPath+"/messages", "tok-bob", nil); code != 403 {
		t.Fatalf("non-member history: %d", code)
	}

	aliceWS, _, err := dial(t, srv, room.ID, "tok-alice")
	if err != nil {
		t.Fatal(err)
	}

	if code, _ := call(t, srv, "POST", roomPath+"/members", "tok-alice", map[string]string{"username": "bob"}); code != 201 {
		t.Fatalf("add bob: %d", code)
	}
	if code, _ := call(t, srv, "POST", roomPath+"/members", "tok-alice", map[string]string{"username": "bob"}); code != 409 {
		t.Fatalf("duplicate: %d", code)
	}
	if e := next(t, aliceWS); e.Type != model.EventMemberAdded || e.Member.Username != "bob" {
		t.Fatalf("alice saw %+v", e)
	}

	bobWS, _, err := dial(t, srv, room.ID, "tok-bob")
	if err != nil {
		t.Fatal(err)
	}

	if err := bobWS.WriteJSON(map[string]string{"content": "hello alice"}); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]*websocket.Conn{"alice": aliceWS, "bob": bobWS} {
		e := next(t, c)
		if e.Type != model.EventMessage || e.Message.Content != "hello alice" || e.Message.Username != "bob" {
			t.Fatalf("%s saw %+v", name, e)
		}
	}

	code, body = call(t, srv, "GET", roomPath+"/messages", "tok-alice", nil)
	var hist []model.Message
	_ = json.Unmarshal(body, &hist)
	if code != 200 || len(hist) != 1 || hist[0].Content != "hello alice" {
		t.Fatalf("history: %d %s", code, body)
	}

	// invalid frames get an error back and keep the connection
	_ = bobWS.WriteJSON(map[string]string{"content": "   "})
	_ = bobWS.SetReadDeadline(time.Now().Add(3 * time.Second))
	var errFrame map[string]string
	if err := bobWS.ReadJSON(&errFrame); err != nil || errFrame["type"] != "error" {
		t.Fatalf("error frame: %v %v", errFrame, err)
	}

	// alice removes bob: he is told, then disconnected
	if code, _ := call(t, srv, "DELETE", roomPath+"/members/bob", "tok-alice", nil); code != 204 {
		t.Fatalf("remove bob: %d", code)
	}
	if e := next(t, bobWS); e.Type != model.EventMemberRemoved {
		t.Fatalf("bob saw %+v", e)
	}
	_ = bobWS.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, _, err := bobWS.ReadMessage(); err == nil {
		t.Fatal("bob should be disconnected")
	}
}
