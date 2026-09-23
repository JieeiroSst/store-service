package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

type ReferralParams struct {
	fx.In

	Referrals port.ReferralRepository
	Gateway   port.ReferralGateway
	Trades    port.TradeRepository
	Notifier  port.Notifier
	Opts      Options
}

type referralService struct {
	referrals port.ReferralRepository
	gateway   port.ReferralGateway
	trades    port.TradeRepository
	notifier  port.Notifier
	opts      Options
}

func NewReferralService(p ReferralParams) *referralService {
	return &referralService{referrals: p.Referrals, gateway: p.Gateway, trades: p.Trades, notifier: p.Notifier, opts: p.Opts.withDefaults()}
}

func (s *referralService) Summary(ctx context.Context, userID string) (*port.ReferralSummary, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	code, err := s.referrals.GetCode(ctx, userID)
	if errors.Is(err, port.ErrNotFound) {
		link, gerr := s.gateway.GenerateLink(ctx, userID)
		if gerr != nil {
			return nil, gerr
		}
		code = &model.ReferralCode{UserID: userID, RefCode: link.RefCode, DeepLink: link.DeepLink, CreatedAt: time.Now()}
		if err := s.referrals.SaveCode(ctx, code); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	out := &port.ReferralSummary{RefCode: code.RefCode, DeepLink: code.DeepLink, CommissionBps: s.commissionBps()}
	if r, err := s.referrals.GetReferral(ctx, userID); err == nil {
		out.ReferredBy = r.ReferrerUserID
	} else if !errors.Is(err, port.ErrNotFound) {
		return nil, err
	}
	if out.Referees, err = s.referrals.CountReferees(ctx, userID); err != nil {
		return nil, err
	}
	if out.Earnings, err = s.referrals.Earnings(ctx, userID); err != nil {
		return nil, err
	}
	if stats, err := s.gateway.Stats(ctx, userID); err == nil {
		out.ServiceStats = stats
	}
	return out, nil
}

func (s *referralService) commissionBps() int64 { return s.opts.ReferralBps }

func (s *referralService) Redeem(ctx context.Context, userID, refCode string) (*model.Referral, error) {
	if userID == "" {
		return nil, port.ErrInvalidUser
	}
	if refCode == "" {
		return nil, fmt.Errorf("%w: ref_code is required", port.ErrInvalidInput)
	}
	if _, err := s.referrals.GetReferral(ctx, userID); err == nil {
		return nil, port.ErrReferralNotAllowed
	} else if !errors.Is(err, port.ErrNotFound) {
		return nil, err
	}
	traded, err := s.trades.ListByUser(ctx, userID, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(traded) > 0 {
		return nil, port.ErrReferralNotAllowed
	}

	owner, err := s.gateway.Activate(ctx, refCode, userID)
	if err != nil {
		return nil, err
	}
	if owner == "" || owner == userID {
		return nil, port.ErrReferralNotAllowed
	}

	r := &model.Referral{RefereeUserID: userID, ReferrerUserID: owner, RefCode: refCode, CreatedAt: time.Now()}
	if err := s.referrals.SaveReferral(ctx, r); err != nil {
		if errors.Is(err, port.ErrAlreadyExists) {
			return nil, port.ErrReferralNotAllowed
		}
		return nil, err
	}
	notify(context.WithoutCancel(ctx), s.notifier, port.Notification{
		UserID: owner, Title: "New referral", Message: "Someone joined with your referral code. You earn a commission on their trading fees.",
	})
	return r, nil
}
