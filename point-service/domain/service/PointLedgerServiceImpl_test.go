package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/domain/entity"
)

type fakeLedger struct {
	credited []entity.PointTransaction
	limit    int
}

func (f *fakeLedger) Credit(_ context.Context, tx *entity.PointTransaction) (bool, error) {
	f.credited = append(f.credited, *tx)
	return true, nil
}

func (f *fakeLedger) GetAccount(_ context.Context, userID string) (*entity.PointAccount, error) {
	return &entity.PointAccount{UserID: userID}, nil
}

func (f *fakeLedger) ListTransactions(_ context.Context, _ string, limit, _ int) ([]entity.PointTransaction, int64, error) {
	f.limit = limit
	return nil, 0, nil
}

func TestApplyValidates(t *testing.T) {
	f := &fakeLedger{}
	s := NewPointLedgerServiceImpl(f)
	ctx := context.Background()
	bad := []dto.RewardEventDTO{
		{UserID: "1", Points: 1},
		{EventID: "e", Points: 1},
		{EventID: "e", UserID: "1"},
		{EventID: "e", UserID: "1", Points: -5},
		{EventID: "e", UserID: "1", Points: 10_000_001},
		{EventID: strings.Repeat("x", 65), UserID: "1", Points: 1},
	}
	for _, ev := range bad {
		if _, err := s.Apply(ctx, ev); !errors.Is(err, ErrInvalidRewardEvent) {
			t.Fatalf("%+v must be rejected: %v", ev, err)
		}
	}
	ok, err := s.Apply(ctx, dto.RewardEventDTO{EventID: " lucky-play-1 ", UserID: "7", Points: 50, Source: strings.Repeat("s", 40)})
	if err != nil || !ok || len(f.credited) != 1 {
		t.Fatalf("valid event: %v %v", ok, err)
	}
	got := f.credited[0]
	if got.EventID != "lucky-play-1" || len(got.Source) != 32 || got.OccurredAt.IsZero() {
		t.Fatalf("normalised: %+v", got)
	}
	if _, err := s.Transactions(ctx, "7", 1000, -1); err != nil || f.limit != 100 {
		t.Fatalf("page size is capped: %d %v", f.limit, err)
	}
}
