package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
)

type recorder struct{ events []model.Event }

func (r *recorder) Publish(_ context.Context, e model.Event) error {
	r.events = append(r.events, e)
	return nil
}

var (
	alice = model.User{ID: 1, Username: "alice"}
	bob   = model.User{ID: 2, Username: "bob"}
	carol = model.User{ID: 3, Username: "carol"}
)

func setup() (*RoomService, *ChatService, *recorder) {
	store := memory.NewStore()
	rec := &recorder{}
	cfg := &config.Config{}
	cfg.Chat.MaxMessageLen = 10
	cfg.Chat.HistoryLimit = 3
	return NewRoomService(store, rec), NewChatService(store, memory.Messages{Store: store}, rec, cfg), rec
}

func mustErr(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("err = %v, want %v", got, want)
	}
}

func TestOwnerManagesMembers(t *testing.T) {
	ctx := context.Background()
	rooms, _, rec := setup()

	room, err := rooms.CreateRoom(ctx, alice, "  general ")
	if err != nil || room.Name != "general" || room.OwnerName != "alice" || room.MemberCount != 1 {
		t.Fatalf("CreateRoom = %+v, %v", room, err)
	}

	if _, err := rooms.AddMember(ctx, alice, room.ID, "bob"); err != nil {
		t.Fatal(err)
	}
	if len(rec.events) != 1 || rec.events[0].Type != model.EventMemberAdded {
		t.Fatalf("events = %+v", rec.events)
	}
	_, err = rooms.AddMember(ctx, alice, room.ID, "bob")
	mustErr(t, err, port.ErrAlreadyMember)
	_, err = rooms.AddMember(ctx, bob, room.ID, "carol") // not owner
	mustErr(t, err, port.ErrForbidden)
	_, err = rooms.AddMember(ctx, alice, room.ID, "bad name")
	mustErr(t, err, port.ErrInvalidInput)

	if list, _ := rooms.ListRooms(ctx, bob); len(list) != 1 || list[0].MemberCount != 2 {
		t.Fatalf("bob's rooms = %+v", list)
	}
	if list, _ := rooms.ListRooms(ctx, carol); len(list) != 0 {
		t.Fatalf("carol's rooms = %+v", list)
	}
}

func TestRemoveMemberRules(t *testing.T) {
	ctx := context.Background()
	rooms, _, rec := setup()
	room, _ := rooms.CreateRoom(ctx, alice, "r")
	_, _ = rooms.AddMember(ctx, alice, room.ID, "bob")
	_, _ = rooms.AddMember(ctx, alice, room.ID, "carol")

	mustErr(t, rooms.RemoveMember(ctx, bob, room.ID, "carol"), port.ErrForbidden) // not owner, not self
	mustErr(t, rooms.RemoveMember(ctx, alice, room.ID, "alice"), port.ErrForbidden)
	mustErr(t, rooms.RemoveMember(ctx, alice, room.ID, "nobody"), port.ErrNotFound)

	if err := rooms.RemoveMember(ctx, bob, room.ID, "bob"); err != nil { // leave
		t.Fatal(err)
	}
	if err := rooms.RemoveMember(ctx, alice, room.ID, "carol"); err != nil { // kick
		t.Fatal(err)
	}
	last := rec.events[len(rec.events)-1]
	if last.Type != model.EventMemberRemoved || last.Member.Username != "carol" {
		t.Fatalf("last event = %+v", last)
	}
	_, err := rooms.ListMembers(ctx, bob, room.ID)
	mustErr(t, err, port.ErrForbidden)
}

func TestNonMemberAndMissingRoom(t *testing.T) {
	ctx := context.Background()
	rooms, chat, _ := setup()
	room, _ := rooms.CreateRoom(ctx, alice, "r")

	_, err := rooms.GetRoom(ctx, bob, room.ID)
	mustErr(t, err, port.ErrForbidden)
	_, err = rooms.GetRoom(ctx, bob, 999)
	mustErr(t, err, port.ErrNotFound)
	_, err = chat.Send(ctx, bob, room.ID, "hi")
	mustErr(t, err, port.ErrForbidden)
	mustErr(t, chat.Authorize(ctx, bob, 999), port.ErrNotFound)
}

func TestChat(t *testing.T) {
	ctx := context.Background()
	rooms, chat, rec := setup()
	room, _ := rooms.CreateRoom(ctx, alice, "r")

	_, err := chat.Send(ctx, alice, room.ID, "   ")
	mustErr(t, err, port.ErrInvalidInput)
	_, err = chat.Send(ctx, alice, room.ID, "this is far too long")
	mustErr(t, err, port.ErrInvalidInput)

	for _, text := range []string{"1", "2", "3", "4"} {
		if _, err := chat.Send(ctx, alice, room.ID, text); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(rec.events); n != 4 || rec.events[3].Message.Content != "4" || rec.events[3].Message.ID == 0 {
		t.Fatalf("events = %+v", rec.events)
	}

	// limit is capped at HistoryLimit (3), returned oldest first
	hist, _ := chat.History(ctx, alice, room.ID, 0, 100)
	if len(hist) != 3 || hist[0].Content != "2" || hist[2].Content != "4" {
		t.Fatalf("history = %+v", hist)
	}
	older, _ := chat.History(ctx, alice, room.ID, hist[0].ID, 3)
	if len(older) != 1 || older[0].Content != "1" {
		t.Fatalf("older = %+v", older)
	}
}
