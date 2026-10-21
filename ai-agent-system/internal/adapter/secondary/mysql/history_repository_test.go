package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"

	"aiagentsystem/internal/domain/model"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	addr := os.Getenv("AI_AGENT_TEST_MYSQL_ADDR")
	if addr == "" {
		t.Skip("AI_AGENT_TEST_MYSQL_ADDR not set")
	}

	cfg := driver.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = addr
	cfg.User = os.Getenv("AI_AGENT_TEST_MYSQL_USER")
	cfg.Passwd = os.Getenv("AI_AGENT_TEST_MYSQL_PASSWORD")
	cfg.DBName = os.Getenv("AI_AGENT_TEST_MYSQL_DATABASE")
	cfg.ParseTime = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return db
}

func TestHistoryRepositoryRoundTrip(t *testing.T) {
	db := openTestDB(t)
	repo := NewHistoryRepository(db)
	ctx := context.Background()

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate twice: %v", err)
	}

	userID := fmt.Sprintf("test-%d", time.Now().UnixNano())
	t.Cleanup(func() { db.Exec(`DELETE FROM chat_messages WHERE user_id = ?`, userID) })

	for _, q := range []string{"first", "second"} {
		turn := model.Turn{UserID: userID, Question: q, Answer: "re: " + q, Source: model.SourceOllama}
		if err := repo.SaveTurn(ctx, turn); err != nil {
			t.Fatalf("SaveTurn: %v", err)
		}
	}

	entries, err := repo.ListByUser(ctx, userID, 3)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}

	newest := entries[0]
	if newest.Role != model.RoleAssistant || newest.Content != "re: second" || newest.Source != "ollama" || newest.UserID != userID {
		t.Fatalf("newest = %+v", newest)
	}
	if entries[1].Role != model.RoleUser || entries[1].Content != "second" || entries[1].Source != "" {
		t.Fatalf("second newest = %+v", entries[1])
	}
	if entries[0].CreatedAt.IsZero() {
		t.Fatal("created_at not scanned")
	}

	other, err := repo.ListByUser(ctx, userID+"-other", 10)
	if err != nil || len(other) != 0 {
		t.Fatalf("other user: %v, %v", other, err)
	}
}
