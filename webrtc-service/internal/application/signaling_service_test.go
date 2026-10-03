package application

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
	"github.com/JIeeiroSst/webrtc-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/webrtc-service/internal/domain"
)

type fakePeer struct {
	id, room string
	inbox    []domain.Message
	full     bool
	closed   bool
	media    domain.MediaState
}

func (p *fakePeer) Media() domain.MediaState     { return p.media }
func (p *fakePeer) SetMedia(m domain.MediaState) { p.media = m }

func (p *fakePeer) ID() string     { return p.id }
func (p *fakePeer) RoomID() string { return p.room }
func (p *fakePeer) Close()         { p.closed = true }

func (p *fakePeer) Send(msg domain.Message) error {
	if p.closed {
		return domain.ErrPeerClosed
	}
	if p.full {
		return domain.ErrPeerBusy
	}
	p.inbox = append(p.inbox, msg)
	return nil
}

func (p *fakePeer) last(t *testing.T) domain.Message {
	t.Helper()
	if len(p.inbox) == 0 {
		t.Fatalf("peer %s received nothing", p.id)
	}
	return p.inbox[len(p.inbox)-1]
}

type countingMetrics struct {
	relayed  map[domain.MessageType]int
	rejected map[string]int
}

func (m *countingMetrics) MessageRelayed(t domain.MessageType) { m.relayed[t]++ }
func (m *countingMetrics) JoinRejected(reason string)          { m.rejected[reason]++ }

func newServiceWithMetrics(maxPeers int) (*SignalingService, *countingMetrics) {
	m := &countingMetrics{relayed: map[domain.MessageType]int{}, rejected: map[string]int{}}
	cfg := &config.Config{Room: config.RoomConfig{MaxPeers: maxPeers}}
	svc := NewSignalingService(memory.NewRoomRegistry(), m, cfg).(*SignalingService)
	svc.now = func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }
	return svc, m
}

func newService() *SignalingService {
	svc, _ := newServiceWithMetrics(0)
	return svc
}

func (p *fakePeer) ofType(t domain.MessageType) []domain.Message {
	var out []domain.Message
	for _, m := range p.inbox {
		if m.Type == t {
			out = append(out, m)
		}
	}
	return out
}

func TestJoinSendsRoomUsersAndNotifiesOthers(t *testing.T) {
	svc := newService()
	a := &fakePeer{id: "a", room: "r"}
	b := &fakePeer{id: "b", room: "r"}

	if err := svc.Join(a); err != nil {
		t.Fatal(err)
	}
	if err := svc.Join(b); err != nil {
		t.Fatal(err)
	}

	got := b.ofType(domain.MessageRoomUsers)[0]
	var users []string
	json.Unmarshal(got.Data, &users)
	if got.Type != domain.MessageRoomUsers || len(users) != 1 || users[0] != "a" {
		t.Fatalf("b got %+v (users %v), want room_users [a]", got, users)
	}
	if got := a.last(t); got.Type != domain.MessageUserJoined || got.UserID != "b" {
		t.Fatalf("a got %+v, want user_joined b", got)
	}
}

func TestJoinRejectsInvalidIDs(t *testing.T) {
	svc := newService()
	for _, p := range []*fakePeer{{id: "", room: "r"}, {id: "a", room: "bad\nroom"}} {
		if err := svc.Join(p); !errors.Is(err, domain.ErrInvalidID) {
			t.Errorf("Join(%q, %q) = %v, want ErrInvalidID", p.id, p.room, err)
		}
	}
}

func TestJoinWithSameIDReplacesOldSession(t *testing.T) {
	svc := newService()
	old := &fakePeer{id: "a", room: "r"}
	b := &fakePeer{id: "b", room: "r"}
	fresh := &fakePeer{id: "a", room: "r"}
	svc.Join(old)
	svc.Join(b)
	b.inbox = nil

	svc.Join(fresh)

	if !old.closed {
		t.Fatal("old session was not closed")
	}
	var reason domain.ErrorPayload
	json.Unmarshal(old.last(t).Data, &reason)
	if reason.Code != "replaced" {
		t.Fatalf("old session got %+v, want error code replaced", old.last(t))
	}
	if len(b.inbox) != 2 || b.inbox[0].Type != domain.MessageUserLeft || b.inbox[1].Type != domain.MessageUserJoined {
		t.Fatalf("b got %+v, want user_left then user_joined", b.inbox)
	}

	b.inbox = nil
	svc.Leave(old)
	if len(b.inbox) != 0 {
		t.Fatalf("Leave(old) notified b: %+v", b.inbox)
	}
	if err := svc.Relay(b, domain.Message{Type: domain.MessageOffer, ToUserID: "a"}); err != nil {
		t.Fatalf("relay to fresh session: %v", err)
	}
}

func TestLeaveNotifiesRemainingMembers(t *testing.T) {
	svc := newService()
	a := &fakePeer{id: "a", room: "r"}
	b := &fakePeer{id: "b", room: "r"}
	svc.Join(a)
	svc.Join(b)

	svc.Leave(b)

	if got := a.last(t); got.Type != domain.MessageUserLeft || got.UserID != "b" {
		t.Fatalf("a got %+v, want user_left b", got)
	}
}

func TestRelayDeliversOnlyToTargetWithSenderIdentity(t *testing.T) {
	svc := newService()
	a := &fakePeer{id: "a", room: "r"}
	b := &fakePeer{id: "b", room: "r"}
	c := &fakePeer{id: "c", room: "r"}
	svc.Join(a)
	svc.Join(b)
	svc.Join(c)
	a.inbox, c.inbox = nil, nil

	sdp := json.RawMessage(`{"type":"offer","sdp":"v=0"}`)
	err := svc.Relay(b, domain.Message{Type: domain.MessageOffer, UserID: "spoofed", RoomID: "other", ToUserID: "a", Data: sdp})
	if err != nil {
		t.Fatal(err)
	}

	got := a.last(t)
	if got.UserID != "b" || got.RoomID != "r" || string(got.Data) != string(sdp) {
		t.Fatalf("a got %+v, want offer from b in r with the original data", got)
	}
	if len(c.inbox) != 0 {
		t.Fatalf("c received %+v, want nothing", c.inbox)
	}
}

func TestRelayRejects(t *testing.T) {
	svc := newService()
	a := &fakePeer{id: "a", room: "r"}
	elsewhere := &fakePeer{id: "x", room: "other"}
	svc.Join(a)
	svc.Join(elsewhere)

	cases := []struct {
		name string
		msg  domain.Message
		want error
	}{
		{"server-only type", domain.Message{Type: domain.MessageUserJoined, ToUserID: "x"}, domain.ErrInvalidMessage},
		{"unknown type", domain.Message{Type: "kick", ToUserID: "x"}, domain.ErrInvalidMessage},
		{"missing target", domain.Message{Type: domain.MessageOffer}, domain.ErrInvalidMessage},
		{"to self", domain.Message{Type: domain.MessageOffer, ToUserID: "a"}, domain.ErrInvalidMessage},
		{"other room", domain.Message{Type: domain.MessageOffer, ToUserID: "x"}, domain.ErrPeerNotFound},
	}
	for _, tc := range cases {
		if err := svc.Relay(a, tc.msg); !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if len(elsewhere.inbox) != 1 {
		t.Fatalf("peer in another room got %+v", elsewhere.inbox)
	}
}

func TestSlowPeerIsDisconnected(t *testing.T) {
	svc := newService()
	a := &fakePeer{id: "a", room: "r"}
	b := &fakePeer{id: "b", room: "r"}
	svc.Join(a)
	svc.Join(b)
	a.full = true

	err := svc.Relay(b, domain.Message{Type: domain.MessageAnswer, ToUserID: "a"})

	if !errors.Is(err, domain.ErrPeerBusy) || !a.closed {
		t.Fatalf("err = %v, closed = %v; want ErrPeerBusy and a closed", err, a.closed)
	}
}

func TestJoinRejectsWhenRoomFull(t *testing.T) {
	svc, m := newServiceWithMetrics(2)
	svc.Join(&fakePeer{id: "a", room: "r"})
	svc.Join(&fakePeer{id: "b", room: "r"})

	if err := svc.Join(&fakePeer{id: "c", room: "r"}); !errors.Is(err, domain.ErrRoomFull) {
		t.Fatalf("err = %v, want ErrRoomFull", err)
	}
	if m.rejected["room_full"] != 1 {
		t.Fatalf("rejected = %v, want room_full counted", m.rejected)
	}
}

func TestChatBroadcastAndPrivate(t *testing.T) {
	svc, m := newServiceWithMetrics(0)
	a := &fakePeer{id: "a", room: "r"}
	b := &fakePeer{id: "b", room: "r"}
	c := &fakePeer{id: "c", room: "r"}
	svc.Join(a)
	svc.Join(b)
	svc.Join(c)
	a.inbox, b.inbox, c.inbox = nil, nil, nil

	if err := svc.Relay(a, domain.Message{Type: domain.MessageChat, Data: json.RawMessage(`{"text":"  hi all "}`)}); err != nil {
		t.Fatal(err)
	}
	if len(a.inbox) != 0 || len(b.inbox) != 1 || len(c.inbox) != 1 {
		t.Fatalf("broadcast reached a=%d b=%d c=%d, want 0/1/1", len(a.inbox), len(b.inbox), len(c.inbox))
	}
	var chat domain.ChatPayload
	json.Unmarshal(b.last(t).Data, &chat)
	if chat.Text != "hi all" || !chat.SentAt.Equal(svc.now()) || b.last(t).UserID != "a" {
		t.Fatalf("b got %+v (%+v), want trimmed text from a with server time", b.last(t), chat)
	}

	if err := svc.Relay(a, domain.Message{Type: domain.MessageChat, ToUserID: "c", Data: json.RawMessage(`{"text":"psst"}`)}); err != nil {
		t.Fatal(err)
	}
	if len(b.inbox) != 1 || len(c.inbox) != 2 {
		t.Fatal("private chat leaked or was not delivered")
	}
	if m.relayed[domain.MessageChat] != 2 {
		t.Fatalf("relayed = %v, want 2 chats", m.relayed)
	}

	for _, data := range []string{`{"text":"   "}`, `"just a string"`, `{"text":"` + strings.Repeat("x", domain.MaxChatChars+1) + `"}`} {
		if err := svc.Relay(a, domain.Message{Type: domain.MessageChat, Data: json.RawMessage(data)}); !errors.Is(err, domain.ErrInvalidMessage) {
			t.Errorf("chat %s: err = %v, want ErrInvalidMessage", data, err)
		}
	}
}

func TestMediaStateIsBroadcastAndReplayedToNewcomers(t *testing.T) {
	svc := newService()
	a := &fakePeer{id: "a", room: "r", media: domain.DefaultMediaState()}
	b := &fakePeer{id: "b", room: "r", media: domain.DefaultMediaState()}
	svc.Join(a)
	svc.Join(b)
	b.inbox = nil

	if err := svc.Relay(a, domain.Message{Type: domain.MessageMediaState, Data: json.RawMessage(`{"audio":false,"video":true,"screen":true}`)}); err != nil {
		t.Fatal(err)
	}
	want := domain.MediaState{Audio: false, Video: true, Screen: true}
	if a.media != want {
		t.Fatalf("a.media = %+v, want %+v", a.media, want)
	}
	var got domain.MediaState
	json.Unmarshal(b.last(t).Data, &got)
	if b.last(t).UserID != "a" || got != want {
		t.Fatalf("b got %+v, want a's new media state", b.last(t))
	}

	c := &fakePeer{id: "c", room: "r"}
	svc.Join(c)
	states := c.ofType(domain.MessageMediaState)
	if len(states) != 2 {
		t.Fatalf("newcomer got %d media states, want one per member", len(states))
	}

	info, err := svc.Room("r")
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Participants) != 3 || info.Participants[0].UserID != "a" || info.Participants[0].Media != want {
		t.Fatalf("Room = %+v", info)
	}
	if _, err := svc.Room(""); !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("Room(\"\") err = %v, want ErrInvalidID", err)
	}
}
