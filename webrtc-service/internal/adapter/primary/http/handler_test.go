package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
	"github.com/JIeeiroSst/webrtc-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/webrtc-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/webrtc-service/internal/application"
	"github.com/JIeeiroSst/webrtc-service/internal/domain"
	"github.com/gorilla/websocket"
)

func newTestServer(t *testing.T) (*httptest.Server, *Handler) {
	t.Helper()
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("MAX_PEERS_PER_ROOM", "3")
	t.Setenv("STUN_URLS", "stun:stun.example:3478")
	t.Setenv("TURN_URLS", "turn:turn.example:3478")
	t.Setenv("TURN_SECRET", "s3cret")
	cfg := config.FromEnv()
	rooms := memory.NewRoomRegistry()
	svc := application.NewSignalingService(rooms, metrics.NewPrometheus(rooms), cfg)
	h := NewHandler(svc, application.NewICEService(cfg), cfg)
	srv := httptest.NewServer(NewRouter(h))
	t.Cleanup(func() {
		h.Shutdown()
		srv.Close()
	})
	return srv, h
}

func dial(t *testing.T, srv *httptest.Server, userID, roomID string) *websocket.Conn {
	t.Helper()
	q := url.Values{"user_id": {userID}, "room_id": {roomID}}
	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?" + q.Encode()
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", userID, err)
	}
	t.Cleanup(func() { ws.Close() })
	return ws
}

func read(t *testing.T, ws *websocket.Conn) domain.Message {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	var msg domain.Message
	if err := ws.ReadJSON(&msg); err != nil {
		t.Fatalf("read: %v", err)
	}
	return msg
}

func TestSignalingOverWebSocket(t *testing.T) {
	srv, _ := newTestServer(t)

	a := dial(t, srv, "alice", "room1")
	if got := read(t, a); got.Type != domain.MessageRoomUsers || string(got.Data) != "[]" {
		t.Fatalf("alice got %+v, want empty room_users", got)
	}

	b := dial(t, srv, "bob", "room1")
	if got := read(t, b); got.Type != domain.MessageRoomUsers || string(got.Data) != `["alice"]` {
		t.Fatalf("bob got %+v, want room_users [alice]", got)
	}
	if got := read(t, b); got.Type != domain.MessageMediaState || got.UserID != "alice" {
		t.Fatalf("bob got %+v, want alice's media_state", got)
	}
	if got := read(t, a); got.Type != domain.MessageUserJoined || got.UserID != "bob" {
		t.Fatalf("alice got %+v, want user_joined bob", got)
	}

	b.WriteJSON(map[string]any{"type": "offer", "user_id": "mallory", "to_user_id": "alice", "data": map[string]string{"sdp": "v=0"}})
	got := read(t, a)
	if got.Type != domain.MessageOffer || got.UserID != "bob" || got.RoomID != "room1" || string(got.Data) != `{"sdp":"v=0"}` {
		t.Fatalf("alice got %+v, want offer from bob", got)
	}

	b.WriteMessage(websocket.TextMessage, []byte("not json"))
	if got := read(t, b); got.Type != domain.MessageError {
		t.Fatalf("bob got %+v, want error", got)
	}
	b.WriteJSON(map[string]any{"type": "answer", "to_user_id": "nobody"})
	if got := read(t, b); got.Type != domain.MessageError {
		t.Fatalf("bob got %+v, want error", got)
	}

	b.Close()
	if got := read(t, a); got.Type != domain.MessageUserLeft || got.UserID != "bob" {
		t.Fatalf("alice got %+v, want user_left bob", got)
	}
}

func TestReconnectWithSameIDKicksOldSession(t *testing.T) {
	srv, _ := newTestServer(t)
	a := dial(t, srv, "alice", "room1")
	read(t, a)
	old := dial(t, srv, "bob", "room1")
	read(t, old)
	read(t, old)
	read(t, a)

	fresh := dial(t, srv, "bob", "room1")
	read(t, fresh)
	read(t, fresh)

	if got := read(t, old); got.Type != domain.MessageError || !strings.Contains(string(got.Data), `"code":"replaced"`) {
		t.Fatalf("old session got %+v, want error code replaced", got)
	}
	old.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, _, err := old.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("old session read err = %v, want normal close", err)
	}
	if got := read(t, a); got.Type != domain.MessageUserLeft {
		t.Fatalf("alice got %+v, want user_left for the stale session", got)
	}
	if got := read(t, a); got.Type != domain.MessageUserJoined || got.UserID != "bob" {
		t.Fatalf("alice got %+v, want user_joined bob", got)
	}

	a.WriteJSON(map[string]any{"type": "offer", "to_user_id": "bob"})
	if got := read(t, fresh); got.Type != domain.MessageOffer || got.UserID != "alice" {
		t.Fatalf("fresh bob got %+v, want offer from alice", got)
	}
}

func TestServeWSRejectsMissingIDs(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, err := http.Get(srv.URL + "/ws?user_id=alice")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestShutdownClosesConnections(t *testing.T) {
	srv, h := newTestServer(t)
	a := dial(t, srv, "alice", "room1")
	read(t, a)

	h.Shutdown()

	a.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, _, err := a.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseGoingAway) {
		t.Fatalf("read err = %v, want going-away close", err)
	}
}

func TestIndexAndHealth(t *testing.T) {
	srv, _ := newTestServer(t)
	for path, want := range map[string]int{
		"/":                200,
		"/health":          200,
		"/missing":         404,
		"/api/rooms/empty": 200,
		"/static/vendor/react-18.3.1.production.min.js": 200,
		"/static/vendor/htm-3.1.1.umd.js":               200,
		"/static/nope.js":                               404,
	} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != want {
			t.Errorf("GET %s = %d, want %d", path, resp.StatusCode, want)
		}
	}
}

func TestCheckOrigin(t *testing.T) {
	req := func(origin string) *http.Request {
		r := httptest.NewRequest("GET", "/ws", nil)
		r.Header.Set("Origin", origin)
		return r
	}
	if checkOrigin(nil) != nil {
		t.Error("no origins configured should fall back to same-origin check")
	}
	if !checkOrigin([]string{"*"})(req("https://evil.example")) {
		t.Error(`"*" should allow any origin`)
	}
	allow := checkOrigin([]string{"https://app.example"})
	if !allow(req("https://app.example")) || allow(req("https://evil.example")) {
		t.Error("explicit list not enforced")
	}
}

func TestJoinFullRoomGetsErrorThenClose(t *testing.T) {
	srv, _ := newTestServer(t)
	for _, id := range []string{"a", "b", "c"} {
		dial(t, srv, id, "room1")
	}

	d := dial(t, srv, "d", "room1")
	if got := read(t, d); got.Type != domain.MessageError || !strings.Contains(string(got.Data), `"code":"room_full"`) {
		t.Fatalf("d got %+v, want room_full error", got)
	}
	d.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, _, err := d.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("read err = %v, want close", err)
	}
}

func TestChatAndMediaStateOverWebSocket(t *testing.T) {
	srv, _ := newTestServer(t)
	a := dial(t, srv, "alice", "room1")
	read(t, a)
	b := dial(t, srv, "bob", "room1")
	read(t, b)
	read(t, b)
	read(t, a)

	a.WriteJSON(map[string]any{"type": "chat", "data": map[string]string{"text": "hello"}})
	got := read(t, b)
	var chat domain.ChatPayload
	json.Unmarshal(got.Data, &chat)
	if got.Type != domain.MessageChat || got.UserID != "alice" || chat.Text != "hello" || chat.SentAt.IsZero() {
		t.Fatalf("bob got %+v, want chat from alice", got)
	}

	b.WriteJSON(map[string]any{"type": "media_state", "data": map[string]bool{"audio": false, "video": true, "screen": false}})
	if got := read(t, a); got.Type != domain.MessageMediaState || got.UserID != "bob" || string(got.Data) != `{"audio":false,"video":true,"screen":false}` {
		t.Fatalf("alice got %+v, want bob muted", got)
	}

	resp, err := http.Get(srv.URL + "/api/rooms/room1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var info domain.RoomInfo
	json.NewDecoder(resp.Body).Decode(&info)
	if info.Capacity != 3 || len(info.Participants) != 2 || info.Participants[1].UserID != "bob" || info.Participants[1].Media.Audio {
		t.Fatalf("room info = %+v", info)
	}
}

func TestICEServersEndpoint(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, err := http.Get(srv.URL + "/api/ice-servers?user_id=alice")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		ICEServers []domain.ICEServer `json:"ice_servers"`
	}
	json.NewDecoder(resp.Body).Decode(&body)
	if resp.Header.Get("Cache-Control") != "no-store" || len(body.ICEServers) != 2 {
		t.Fatalf("ice servers = %+v", body)
	}
	if turn := body.ICEServers[1]; !strings.HasSuffix(turn.Username, ":alice") || turn.Credential == "" {
		t.Fatalf("turn entry = %+v, want ephemeral credentials for alice", turn)
	}
}
