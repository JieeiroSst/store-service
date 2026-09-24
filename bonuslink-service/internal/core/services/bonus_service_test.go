package services

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/JIeeiroSst/bonuslink-service/internal/core/domain"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/ports"
)

type fakeRepo struct {
	seen map[string]bool
}

func (f *fakeRepo) Record(_ context.Context, r *domain.Reward) (bool, error) {
	if f.seen[r.EventID] {
		return false, nil
	}
	f.seen[r.EventID] = true
	return true, nil
}
func (f *fakeRepo) ListByUser(context.Context, string, int, int) ([]*domain.Reward, error) {
	return nil, nil
}
func (f *fakeRepo) BalancesByUser(context.Context, string) ([]*domain.Balance, error) {
	return nil, nil
}

func TestRecordReward(t *testing.T) {
	svc := NewBonusService(&fakeRepo{seen: map[string]bool{}}, zap.NewNop())
	valid := ports.RecordRewardRequest{EventID: "e1", UserID: "u1", RewardType: domain.RewardPoints, Value: 10}

	r, created, err := svc.RecordReward(context.Background(), valid)
	if err != nil || !created || r.Source != "referral" {
		t.Fatalf("first call: reward=%+v created=%v err=%v", r, created, err)
	}

	if _, created, err = svc.RecordReward(context.Background(), valid); err != nil || created {
		t.Fatalf("duplicate must be a no-op: created=%v err=%v", created, err)
	}

	bad := []ports.RecordRewardRequest{
		{UserID: "u1", RewardType: domain.RewardPoints, Value: 1},
		{EventID: "e2", RewardType: domain.RewardPoints, Value: 1},
		{EventID: "e2", UserID: "u1", RewardType: "GOLD", Value: 1},
		{EventID: "e2", UserID: "u1", RewardType: domain.RewardPoints, Value: 0},
	}
	for i, req := range bad {
		if _, _, err := svc.RecordReward(context.Background(), req); !errors.Is(err, domain.ErrInvalidReward) {
			t.Errorf("case %d: want ErrInvalidReward, got %v", i, err)
		}
	}
}
