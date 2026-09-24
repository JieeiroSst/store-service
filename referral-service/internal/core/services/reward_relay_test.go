package services

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
)

type relayRepo struct {
	ports.RewardRepository
	pending []*domain.ReferralReward
	marked  []string
}

func (r *relayRepo) FindUnpublished(context.Context, int64, int) ([]*domain.ReferralReward, error) {
	return r.pending, nil
}
func (r *relayRepo) MarkPublished(_ context.Context, _, refCode string, _ int64) error {
	r.marked = append(r.marked, refCode)
	return nil
}

type relayPublisher struct {
	failOn string
	sent   []ports.RewardGrantedEvent
}

func (p *relayPublisher) PublishRewardGranted(_ context.Context, ev ports.RewardGrantedEvent) error {
	if ev.RefCode == p.failOn {
		return errors.New("broker down")
	}
	p.sent = append(p.sent, ev)
	return nil
}

func TestRelayFlush(t *testing.T) {
	repo := &relayRepo{pending: []*domain.ReferralReward{
		{OwnerUserID: "o", RefCode: "a", NewUserID: "n1", RewardType: domain.RewardCredit, RewardValue: 5},
		{OwnerUserID: "o", RefCode: "b", NewUserID: "n2", RewardType: domain.RewardCoupon, RewardValue: 1},
		{OwnerUserID: "o", RefCode: "c", NewUserID: "n3", RewardType: domain.RewardCredit, RewardValue: 2},
	}}
	pub := &relayPublisher{failOn: "b"}
	rl := &rewardRelay{rewards: repo, publisher: pub, log: zap.NewNop()}

	rl.flush(context.Background())

	// "a" is sent and marked; the failure on "b" stops the batch, leaving b and c for the next tick.
	if len(pub.sent) != 1 || pub.sent[0].EventID != "referral:a:n1" || pub.sent[0].RewardType != "POINTS" {
		t.Fatalf("sent = %+v", pub.sent)
	}
	if len(repo.marked) != 1 || repo.marked[0] != "a" {
		t.Fatalf("marked = %v", repo.marked)
	}
}
