package postgres

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	server := os.Getenv("EJABBERD_TEST_SQL_SERVER")
	if server == "" {
		t.Skip("EJABBERD_TEST_SQL_SERVER not set")
	}
	cfg := config.Defaults()
	cfg.SQL.Server = server
	if p, err := strconv.Atoi(os.Getenv("EJABBERD_TEST_SQL_PORT")); err == nil {
		cfg.SQL.Port = p
	}
	if v := os.Getenv("EJABBERD_TEST_SQL_PASSWORD"); v != "" {
		cfg.SQL.Password = v
	}
	if v := os.Getenv("EJABBERD_TEST_SQL_USERNAME"); v != "" {
		cfg.SQL.Username = v
	}
	if v := os.Getenv("EJABBERD_TEST_SQL_DATABASE"); v != "" {
		cfg.SQL.Database = v
	}
	db, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := db.Migrate(ctx, true); err != nil {
		t.Fatal(err)
	}
	for _, tbl := range []string{"users", "rosterusers", "rostergroups", "roster_version", "spool"} {
		if _, err := db.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUserStore(t *testing.T) {
	ctx := context.Background()
	s := NewUserStore(testDB(t))
	if err := s.TryRegister(ctx, "alice", "localhost", domain.Password{Plain: "pw"}); err != nil {
		t.Fatal(err)
	}
	if err := s.TryRegister(ctx, "alice", "localhost", domain.Password{Plain: "pw"}); err != domain.ErrUserExists {
		t.Fatalf("duplicate = %v", err)
	}
	sc := domain.PasswordToScramWithSalt(domain.ScramSHA256, "secret", []byte("0123456789abcdef"), 4096)
	if err := s.SetPassword(ctx, "alice", "localhost", domain.Password{Scram: sc}); err != nil {
		t.Fatal(err)
	}
	p, err := s.GetPassword(ctx, "alice", "localhost")
	if err != nil || !p.IsScram() || p.Scram.Hash != domain.ScramSHA256 || !domain.MatchPassword("secret", p) {
		t.Fatalf("scram round trip: %+v %v", p, err)
	}
	if n, _ := s.CountUsers(ctx, "localhost"); n != 1 {
		t.Fatalf("count = %d", n)
	}
	if err := s.RemoveUser(ctx, "alice", "localhost"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetPassword(ctx, "alice", "localhost"); err != domain.ErrUserNotFound {
		t.Fatalf("after remove = %v", err)
	}
}

func TestRosterStore(t *testing.T) {
	ctx := context.Background()
	s := NewRosterStore(testDB(t))
	bob := domain.LJID{User: "bob", Server: "localhost"}
	err := s.Transaction(ctx, "alice", "localhost", func(tx port.RosterRepository) error {
		return tx.UpdateRoster(ctx, domain.RosterItem{
			LUser: "alice", LServer: "localhost", JID: bob, Name: "Bob",
			Subscription: domain.SubTo, Ask: domain.AskIn, Groups: []string{"Friends", "Work"}, AskMessage: "hi",
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	it, ok, err := s.GetRosterItem(ctx, "alice", "localhost", bob)
	if err != nil || !ok || it.Subscription != domain.SubTo || it.Ask != domain.AskIn || len(it.Groups) != 2 || it.AskMessage != "hi" {
		t.Fatalf("item = %+v %v %v", it, ok, err)
	}
	if err := s.DeleteRoster(ctx, "alice", "localhost", bob); err != nil {
		t.Fatal(err)
	}
	if items, _ := s.GetRoster(ctx, "alice", "localhost"); len(items) != 0 {
		t.Fatalf("roster after delete = %+v", items)
	}
}

func TestOfflineStore(t *testing.T) {
	ctx := context.Background()
	s := NewOfflineStore(testDB(t))
	from := domain.MustJID("bob", "localhost", "r")
	to := domain.MustJID("alice", "localhost", "")
	ts := time.Now().Add(-time.Minute).Truncate(time.Microsecond)
	for _, body := range []string{"one", "two"} {
		msg := domain.NewMessage("chat", from, to, "", body)
		if err := s.Store(ctx, domain.OfflineMessage{LUser: "alice", LServer: "localhost", Timestamp: ts, From: from, To: to, Packet: msg}); err != nil {
			t.Fatal(err)
		}
	}
	if n, _ := s.Count(ctx, "alice", "localhost"); n != 2 {
		t.Fatalf("count = %d", n)
	}
	msgs, err := s.Pop(ctx, "alice", "localhost")
	if err != nil || len(msgs) != 2 {
		t.Fatalf("pop = %d %v", len(msgs), err)
	}
	if msgs[0].Packet.ChildText("body", "") != "one" || !msgs[0].Timestamp.Equal(ts) || msgs[0].Packet.Child("delay", domain.NSDelay) != nil {
		t.Fatalf("first = %s ts=%v", msgs[0].Packet, msgs[0].Timestamp)
	}
	if n, _ := s.Count(ctx, "alice", "localhost"); n != 0 {
		t.Fatalf("count after pop = %d", n)
	}
}
