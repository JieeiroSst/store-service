package persistence

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ledgerDB(t *testing.T) *PointLedgerRepositoryImpl {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=busy_timeout(5000)", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.PointAccount{}, &entity.PointTransaction{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return NewPointLedgerRepositoryImpl(db)
}

func credit(id, user string, points int64) *entity.PointTransaction {
	return &entity.PointTransaction{EventID: id, UserID: user, Source: "lucky-wheel", Title: "x", Points: points, OccurredAt: time.Now()}
}

func TestCreditIsIdempotent(t *testing.T) {
	r := ledgerDB(t)
	ctx := context.Background()
	applied, err := r.Credit(ctx, credit("lucky-play-1", "7", 100))
	if err != nil || !applied {
		t.Fatalf("first credit: %v %v", applied, err)
	}
	tx := credit("lucky-play-1", "7", 100)
	applied, err = r.Credit(ctx, tx)
	if err != nil || applied {
		t.Fatalf("redelivered event must be skipped: %v %v", applied, err)
	}
	second := credit("lucky-checkin-1", "7", 20)
	if applied, err := r.Credit(ctx, second); err != nil || !applied || second.BalanceAfter != 120 {
		t.Fatalf("second event: %v %v %+v", applied, err, second)
	}
	if _, err := r.Credit(ctx, credit("lucky-play-2", "8", 5)); err != nil {
		t.Fatal(err)
	}
	a, err := r.GetAccount(ctx, "7")
	if err != nil || a.Balance != 120 || a.TotalEarned != 120 {
		t.Fatalf("account: %+v %v", a, err)
	}
	empty, err := r.GetAccount(ctx, "404")
	if err != nil || empty.Balance != 0 || empty.UserID != "404" {
		t.Fatalf("unknown user has an empty account: %+v %v", empty, err)
	}
	items, total, err := r.ListTransactions(ctx, "7", 10, 0)
	if err != nil || total != 2 || items[0].EventID != "lucky-checkin-1" || items[1].BalanceAfter != 100 {
		t.Fatalf("transactions: %+v %d %v", items, total, err)
	}
}

func TestConcurrentRedelivery(t *testing.T) {
	r := ledgerDB(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	var mu sync.Mutex
	applied := 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ok, err := r.Credit(ctx, credit(fmt.Sprintf("ev-%d", i%3), "9", 10))
			if err != nil {
				t.Errorf("credit: %v", err)
				return
			}
			if ok {
				mu.Lock()
				applied++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	a, _ := r.GetAccount(ctx, "9")
	if applied != 3 || a.Balance != 30 {
		t.Fatalf("each event credited once: applied=%d balance=%d", applied, a.Balance)
	}
}
