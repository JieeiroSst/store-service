package mysql

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/referral/service/internal/core/domain"
)

// Temporary end-to-end verification against the docker-compose MySQL instance.
// Not part of the deliverable test suite — deleted after manual verification.
func TestVerifyMySQLAdapters(t *testing.T) {
	db, err := sqlx.Connect("mysql", "root:root@tcp(localhost:3306)/referral_service?parseTime=true&loc=UTC&charset=utf8mb4")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	logger := zap.NewNop()
	ctx := context.Background()

	linkRepo := NewReferralLinkRepo(db, logger)
	eventRepo := NewReferralEventRepo(db, logger)
	rewardRepo := NewRewardRepo(db, logger)
	statsRepo := NewUserStatsRepo(db, logger)
	programRepo := NewRewardProgramRepo(db, logger)
	attrRepo := NewAttributionRepo(db, logger)

	now := time.Now().UnixMilli()
	refCode := fmt.Sprintf("TST%d", now%1000000)
	owner := "owner-verify-1"

	link := &domain.ReferralLink{
		RefCode: refCode, CreatedAt: now, OwnerUserID: owner, Channel: domain.ChannelCopy,
		Status: domain.StatusActive, ExpiresAt: now + 1000000, DeepLink: "https://example.com/r/" + refCode, Platform: "ios",
	}
	if err := linkRepo.Save(ctx, link); err != nil {
		t.Fatalf("link.Save: %v", err)
	}
	if err := linkRepo.Save(ctx, link); err != domain.ErrDuplicateRefCode {
		t.Fatalf("expected ErrDuplicateRefCode, got %v", err)
	}

	got, err := linkRepo.FindByRefCode(ctx, refCode)
	if err != nil || got.OwnerUserID != owner {
		t.Fatalf("FindByRefCode: err=%v got=%+v", err, got)
	}

	count, err := linkRepo.CountByOwnerUserIDSince(ctx, owner, now-1000)
	if err != nil || count != 1 {
		t.Fatalf("CountByOwnerUserIDSince: err=%v count=%d", err, count)
	}

	for i := 0; i < 5; i++ {
		c := fmt.Sprintf("PG%d%d", now%100000, i)
		if err := linkRepo.Save(ctx, &domain.ReferralLink{
			RefCode: c, CreatedAt: now + int64(i), OwnerUserID: owner, Channel: domain.ChannelCopy,
			Status: domain.StatusActive, ExpiresAt: now + 1000000, DeepLink: "x", Platform: "ios",
		}); err != nil {
			t.Fatalf("seed link: %v", err)
		}
	}

	page1, cursor1, err := linkRepo.FindByOwnerUserID(ctx, owner, 3, "")
	if err != nil || len(page1) != 3 || cursor1 == "" {
		t.Fatalf("page1: err=%v len=%d cursor=%q", err, len(page1), cursor1)
	}
	page2, _, err := linkRepo.FindByOwnerUserID(ctx, owner, 3, cursor1)
	if err != nil || len(page2) == 0 {
		t.Fatalf("page2: err=%v len=%d", err, len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.RefCode == p2.RefCode {
				t.Fatalf("pagination overlap: %s in both pages", p1.RefCode)
			}
		}
	}

	if err := linkRepo.UpdateStatus(ctx, refCode, domain.StatusUsed); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if err := linkRepo.UpdateStatus(ctx, refCode, domain.StatusUsed); err == nil {
		t.Fatalf("expected ErrLinkNotActive on second UpdateStatus call")
	}

	if err := eventRepo.Save(ctx, &domain.ReferralEvent{
		RefCode: refCode, EventID: "evt-1", EventType: domain.EventLinkClicked,
		OccurredAt: now, Platform: "ios", OwnerUserID: owner, DeviceID: "dev-1",
	}); err != nil {
		t.Fatalf("event.Save: %v", err)
	}
	events, err := eventRepo.FindByRefCode(ctx, refCode)
	if err != nil || len(events) != 1 {
		t.Fatalf("event.FindByRefCode: err=%v len=%d", err, len(events))
	}

	if err := rewardRepo.Save(ctx, &domain.ReferralReward{
		OwnerUserID: owner, RefCode: refCode, NewUserID: "newuser-1",
		RewardType: domain.RewardCredit, RewardValue: 50000, Status: domain.RewardPending,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("reward.Save: %v", err)
	}
	reward, err := rewardRepo.FindByOwnerAndRefCode(ctx, owner, refCode)
	if err != nil || reward.RewardValue != 50000 {
		t.Fatalf("reward.FindByOwnerAndRefCode: err=%v reward=%+v", err, reward)
	}

	if err := statsRepo.IncrementCounters(ctx, owner, 1, 0, 0, 0); err != nil {
		t.Fatalf("IncrementCounters first: %v", err)
	}
	if err := statsRepo.IncrementCounters(ctx, owner, 0, 1, 1, 50000); err != nil {
		t.Fatalf("IncrementCounters second: %v", err)
	}
	stats, err := statsRepo.Get(ctx, owner)
	if err != nil || stats.TotalInvited != 1 || stats.TotalInstalled != 1 || stats.TotalRewarded != 1 || stats.TotalRewardAmt != 50000 {
		t.Fatalf("stats mismatch: err=%v stats=%+v", err, stats)
	}

	missingStats, err := statsRepo.Get(ctx, "no-such-user")
	if err != nil || missingStats.TotalInvited != 0 {
		t.Fatalf("expected zero-value stats: err=%v stats=%+v", err, missingStats)
	}

	programID := fmt.Sprintf("prog-%d", now)
	if err := programRepo.Save(ctx, &domain.RewardProgram{
		ProgramID: programID, Name: "Verify Program", Status: domain.ProgramStatusActive,
		Tiers: []domain.RewardTier{
			{MinCount: 1, MaxCount: 9, RewardValue: 50000},
			{MinCount: 10, MaxCount: -1, RewardValue: 100000},
		},
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("program.Save: %v", err)
	}
	fetchedProgram, err := programRepo.FindByID(ctx, programID)
	if err != nil || len(fetchedProgram.Tiers) != 2 || fetchedProgram.Tiers[1].RewardValue != 100000 {
		t.Fatalf("program.FindByID tiers mismatch: err=%v tiers=%+v", err, fetchedProgram)
	}

	active, err := programRepo.FindActive(ctx)
	if err != nil || active == nil || active.ProgramID != programID {
		t.Fatalf("program.FindActive: err=%v active=%+v", err, active)
	}

	if err := attrRepo.Save(ctx, &domain.ReferralAttribution{
		OwnerUserID: owner, NewUserID: "newuser-1", RefCode: refCode, Platform: "ios",
		DeviceID: "dev-1", AttributedAt: now,
	}); err != nil {
		t.Fatalf("attr.Save: %v", err)
	}
	attr, err := attrRepo.FindByNewUserID(ctx, "newuser-1")
	if err != nil || attr.OwnerUserID != owner {
		t.Fatalf("attr.FindByNewUserID: err=%v attr=%+v", err, attr)
	}

	t.Log("ALL MYSQL ADAPTER CHECKS PASSED")
}
