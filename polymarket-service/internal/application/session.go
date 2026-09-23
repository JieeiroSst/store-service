package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

type sessionDeps struct {
	balances  port.BalanceRepository
	positions port.PositionRepository
	referrals port.ReferralRepository
	pnl       port.PnLRepository
	exchange  port.ExchangeAccountRepository
	orders    port.OrderRepository
}

type session struct {
	sessionDeps

	bal       map[string]*model.Balance
	pos       map[string]*model.Position
	referrers map[string]string
	pnlLines  []model.PnLEntry
	exchLines []model.ExchangeEntry
}

func newSession(d sessionDeps) *session {
	return &session{
		sessionDeps: d,
		bal:         map[string]*model.Balance{}, pos: map[string]*model.Position{}, referrers: map[string]string{},
	}
}

func (s *session) balance(ctx context.Context, userID string) (*model.Balance, error) {
	if b, ok := s.bal[userID]; ok {
		return b, nil
	}
	b, err := s.balances.GetForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.bal[userID] = b
	return b, nil
}

func (s *session) position(ctx context.Context, marketID int64, userID string, outcome model.Outcome) (*model.Position, error) {
	key := fmt.Sprintf("%d/%s/%s", marketID, userID, outcome)
	if p, ok := s.pos[key]; ok {
		return p, nil
	}
	p, err := s.positions.GetForUpdate(ctx, marketID, userID, outcome)
	if errors.Is(err, port.ErrNotFound) {
		p = &model.Position{MarketID: marketID, UserID: userID, Outcome: outcome}
	} else if err != nil {
		return nil, err
	}
	s.pos[key] = p
	return p, nil
}

func (s *session) referrer(ctx context.Context, userID string) (string, error) {
	if r, ok := s.referrers[userID]; ok {
		return r, nil
	}
	ref, err := s.referrals.GetReferral(ctx, userID)
	switch {
	case errors.Is(err, port.ErrNotFound):
		s.referrers[userID] = ""
	case err != nil:
		return "", err
	default:
		s.referrers[userID] = ref.ReferrerUserID
	}
	return s.referrers[userID], nil
}

func (s *session) addPnL(userID string, marketID int64, kind model.PnLKind, amount int64) {
	if amount != 0 {
		s.pnlLines = append(s.pnlLines, model.PnLEntry{UserID: userID, MarketID: marketID, Kind: kind, Amount: amount})
	}
}

func (s *session) addExchange(bucket model.ExchangeBucket, kind string, amount, marketID int64, userID string) {
	if amount != 0 {
		s.exchLines = append(s.exchLines, model.ExchangeEntry{Bucket: bucket, Kind: kind, Amount: amount, MarketID: marketID, UserID: userID})
	}
}

func (s *session) flush(ctx context.Context) error {
	for _, b := range s.bal {
		if b.Available < 0 || b.Locked < 0 {
			return fmt.Errorf("invariant violated: balance of %s went negative (%d/%d)", b.UserID, b.Available, b.Locked)
		}
		if err := s.balances.Save(ctx, b); err != nil {
			return err
		}
	}
	for _, p := range s.pos {
		if p.Shares < 0 || p.LockedShares < 0 || p.LockedShares > p.Shares {
			return fmt.Errorf("invariant violated: position %d/%s/%s is inconsistent (%d shares, %d locked)", p.MarketID, p.UserID, p.Outcome, p.Shares, p.LockedShares)
		}
		if err := s.positions.Save(ctx, p); err != nil {
			return err
		}
	}
	if len(s.pnlLines) > 0 {
		if err := s.pnl.CreateMany(ctx, s.pnlLines); err != nil {
			return err
		}
	}
	if len(s.exchLines) > 0 {
		if err := s.exchange.Add(ctx, s.exchLines...); err != nil {
			return err
		}
	}
	return nil
}

func closeOrder(ctx context.Context, sess *session, o *model.Order, status model.OrderStatus) error {
	if o.Side == model.Buy {
		bal, err := sess.balance(ctx, o.UserID)
		if err != nil {
			return err
		}
		held := o.LockedCash + o.FeeReserve
		bal.Locked -= held
		bal.Available += held
		o.LockedCash, o.FeeReserve = 0, 0
	} else if rem := o.Remaining(); rem > 0 {
		pos, err := sess.position(ctx, o.MarketID, o.UserID, o.Outcome)
		if err != nil {
			return err
		}
		pos.LockedShares -= rem
	}
	o.Status = status
	return sess.orders.Save(ctx, o)
}
