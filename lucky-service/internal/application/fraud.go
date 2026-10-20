package application

import (
	"context"
	"log/slog"
	"strings"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"go.uber.org/fx"
)

type Guard struct {
	fraud   port.FraudRepository
	signals port.FraudSignals
	policy  config.FraudConfig
}

type GuardDeps struct {
	fx.In

	Config  *config.Config
	Fraud   port.FraudRepository
	Signals port.FraudSignals `optional:"true"`
}

func NewGuard(d GuardDeps) *Guard {
	return &Guard{fraud: d.Fraud, signals: d.Signals, policy: d.Config.Fraud}
}

func (g *Guard) Screen(ctx context.Context, actor model.Actor, client model.ClientInfo) (string, error) {
	if g == nil {
		return "", nil
	}
	if g.policy.RequireDevice && client.DeviceID == "" {
		return "", model.ErrDeviceRequired
	}
	b, err := g.fraud.BlockOf(ctx, actor.UserID)
	if err != nil {
		return "", err
	}
	if b != nil {
		return "", model.ErrBlocked
	}
	if g.signals == nil {
		return "", nil
	}
	if client.DeviceID != "" && g.policy.MaxUsersPerDevice > 0 {
		n, err := g.signals.Observe(ctx, model.SignalDevice, client.DeviceID, actor.UserID)
		switch {
		case err != nil:
			slog.Warn("fraud signals unavailable, skipping device check", "err", err)
		case n > g.policy.MaxUsersPerDevice:
			slog.Warn("device shared by too many accounts", "user_id", actor.UserID, "device_id", client.DeviceID, "accounts", n)
			return "", model.ErrDeviceLimit
		}
	}
	if client.IP != "" && g.policy.MaxUsersPerIP > 0 {
		n, err := g.signals.Observe(ctx, model.SignalIP, client.IP, actor.UserID)
		switch {
		case err != nil:
			slog.Warn("fraud signals unavailable, skipping ip check", "err", err)
		case n > g.policy.MaxUsersPerIP:
			return model.HoldSharedIP, nil
		}
	}
	return "", nil
}

type FraudService struct {
	fraud port.FraudRepository
	plays port.PlayRepository
}

func NewFraudService(fraud port.FraudRepository, plays port.PlayRepository) port.FraudUsecase {
	return &FraudService{fraud: fraud, plays: plays}
}

func reviewer(actor model.Actor) string {
	if actor.UserID != "" {
		return actor.UserID
	}
	return "service"
}

func (s *FraudService) Block(ctx context.Context, actor model.Actor, in port.BlockInput) (*port.BlockResult, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.Reason = strings.TrimSpace(in.Reason)
	switch {
	case in.UserID == "" || len(in.UserID) > 64:
		return nil, &model.ValidationError{Field: "user_id", Reason: "must be 1-64 characters"}
	case in.Reason == "" || len(in.Reason) > 255:
		return nil, &model.ValidationError{Field: "reason", Reason: "must be 1-255 characters"}
	}
	b := &model.UserBlock{UserID: in.UserID, Reason: in.Reason, BlockedBy: reviewer(actor)}
	held, err := s.fraud.Block(ctx, b)
	if err != nil {
		return nil, err
	}
	slog.Info("user blocked from lucky games", "user_id", b.UserID, "by", b.BlockedBy, "held_rewards", held)
	return &port.BlockResult{Block: *b, HeldRewards: held}, nil
}

func (s *FraudService) Unblock(ctx context.Context, actor model.Actor, userID string) error {
	if err := requireAdmin(actor); err != nil {
		return err
	}
	return s.fraud.Unblock(ctx, strings.TrimSpace(userID))
}

func (s *FraudService) Blocks(ctx context.Context, actor model.Actor, limit, offset int) (*port.BlockPage, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	limit, offset = page(limit, offset)
	items, total, err := s.fraud.Blocks(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return &port.BlockPage{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *FraudService) Rewards(ctx context.Context, actor model.Actor, in port.ListRewardsInput) (*port.RewardPage, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	f := port.RewardFilter{Status: model.RewardStatus(in.Status), UserID: strings.TrimSpace(in.UserID)}
	if f.Status == "" {
		f.Status = model.RewardHeld
	}
	if !f.Status.Valid() {
		return nil, &model.ValidationError{Field: "status", Reason: "must be pending, held or rejected"}
	}
	f.Limit, f.Offset = page(in.Limit, in.Offset)
	items, total, err := s.fraud.Rewards(ctx, f)
	if err != nil {
		return nil, err
	}
	return &port.RewardPage{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

func (s *FraudService) Approve(ctx context.Context, actor model.Actor, id int64) (*model.ReviewReward, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	return s.fraud.Review(ctx, id, []model.RewardStatus{model.RewardHeld}, model.RewardPending, reviewer(actor))
}

func (s *FraudService) Reject(ctx context.Context, actor model.Actor, id int64) (*model.ReviewReward, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	return s.fraud.Review(ctx, id, []model.RewardStatus{model.RewardHeld, model.RewardPending}, model.RewardRejected, reviewer(actor))
}

func (s *FraudService) Plays(ctx context.Context, actor model.Actor, in port.ListPlaysInput) (*port.PlayPage, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	if in.UserID == "" && in.DeviceID == "" && in.IP == "" {
		return nil, &model.ValidationError{Field: "user_id", Reason: "filter by user_id, device_id or ip"}
	}
	limit, offset := page(in.Limit, in.Offset)
	items, total, err := s.plays.List(ctx, port.PlayFilter{UserID: in.UserID, DeviceID: in.DeviceID, IP: in.IP, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return &port.PlayPage{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}
