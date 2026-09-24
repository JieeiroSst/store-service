package mysql

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Needs a live MySQL: ROOM_TEST_MYSQL_DSN='user:pass@tcp(host:port)/db?parseTime=True&loc=UTC'.
func TestRepositories(t *testing.T) {
	dsn := os.Getenv("ROOM_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("ROOM_TEST_MYSQL_DSN not set")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&roomRow{}, &memberRow{}, &messageRow{}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rooms, msgs := NewRoomRepository(db), NewMessageRepository(db)

	room, err := rooms.Create(ctx, "general", "Alice")
	if err != nil || room.OwnerName != "Alice" || room.MemberCount != 1 {
		t.Fatalf("Create = %+v, %v", room, err)
	}

	if err := rooms.AddMember(ctx, model.Member{RoomID: room.ID, Username: "bob", Role: model.RoleMember}); err != nil {
		t.Fatal(err)
	}
	if err := rooms.AddMember(ctx, model.Member{RoomID: room.ID, Username: "bob", Role: model.RoleMember}); !errors.Is(err, port.ErrAlreadyMember) {
		t.Fatalf("duplicate = %v", err)
	}
	// case-sensitive, matching user-service
	if err := rooms.AddMember(ctx, model.Member{RoomID: room.ID, Username: "BOB", Role: model.RoleMember}); err != nil {
		t.Fatalf("BOB should be distinct from bob: %v", err)
	}
	if _, err := rooms.Member(ctx, room.ID, "alice"); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("alice != Alice: %v", err)
	}
	m, err := rooms.Member(ctx, room.ID, "bob")
	if err != nil || m.Role != model.RoleMember || m.JoinedAt.IsZero() {
		t.Fatalf("Member = %+v, %v", m, err)
	}

	list, err := rooms.ListByMember(ctx, "bob")
	found := false
	for _, r := range list {
		if r.ID == room.ID && r.MemberCount == 3 {
			found = true
		}
	}
	if err != nil || !found {
		t.Fatalf("ListByMember = %+v, %v", list, err)
	}
	if err := rooms.RemoveMember(ctx, room.ID, "BOB"); err != nil {
		t.Fatal(err)
	}
	if members, _ := rooms.Members(ctx, room.ID); len(members) != 2 {
		t.Fatalf("members = %+v", members)
	}
	if _, err := rooms.FindByID(ctx, 1<<30); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("missing room = %v", err)
	}

	var ids []uint
	for _, text := range []string{"a", "b", "c"} {
		m := model.Message{RoomID: room.ID, Username: "Alice", Content: text, CreatedAt: time.Now().UTC()}
		if err := msgs.Create(ctx, &m); err != nil || m.ID == 0 {
			t.Fatalf("Create msg: %v", err)
		}
		ids = append(ids, m.ID)
	}
	latest, _ := msgs.List(ctx, room.ID, 0, 2)
	if len(latest) != 2 || latest[0].Content != "b" || latest[1].Content != "c" {
		t.Fatalf("latest = %+v", latest)
	}
	older, _ := msgs.List(ctx, room.ID, ids[1], 5)
	if len(older) != 1 || older[0].Content != "a" {
		t.Fatalf("older = %+v", older)
	}
}
