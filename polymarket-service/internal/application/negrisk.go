package application

import (
	"context"
	"fmt"
	"slices"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func (s *exchangeService) Convert(ctx context.Context, in port.ConvertInput) (*port.ConvertResult, error) {
	if in.UserID == "" {
		return nil, port.ErrInvalidUser
	}
	if in.Amount <= 0 {
		return nil, port.ErrInvalidSize
	}
	give := slices.Clone(in.MarketIDs)
	slices.Sort(give)
	give = slices.Compact(give)
	if len(give) == 0 {
		return nil, fmt.Errorf("%w: market_ids must name at least one market", port.ErrInvalidInput)
	}

	var result *port.ConvertResult
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		event, err := s.events.GetByID(ctx, in.EventID)
		if err != nil {
			return err
		}
		if !event.NegRisk {
			return port.ErrNotNegRisk
		}
		markets, err := s.markets.ListByEventForUpdate(ctx, in.EventID)
		if err != nil {
			return err
		}

		shareValue := markets[0].ShareValue
		giving := map[int64]bool{}
		for _, id := range give {
			giving[id] = true
		}
		for _, m := range markets {
			if !m.Tradable(s.now()) {
				return port.ErrMarketNotTradable
			}
			if m.ShareValue != shareValue {
				return fmt.Errorf("%w: markets of a neg-risk event must share one share_value", port.ErrInvalidInput)
			}
			delete(giving, m.ID)
		}
		if len(giving) > 0 {
			return fmt.Errorf("%w: market_ids must belong to the event", port.ErrInvalidInput)
		}

		sess := s.session()
		k := int64(len(give))
		var released int64
		var receiving []*model.Position
		for _, m := range markets {
			if slices.Contains(give, m.ID) {
				pos, err := sess.position(ctx, m.ID, in.UserID, model.OutcomeNo)
				if err != nil {
					return err
				}
				if pos.Free() < in.Amount {
					return port.ErrInsufficientShares
				}
				cost := pos.CostBasis * in.Amount / pos.Shares
				pos.CostBasis -= cost
				pos.Shares -= in.Amount
				released += cost
				continue
			}
			pos, err := sess.position(ctx, m.ID, in.UserID, model.OutcomeYes)
			if err != nil {
				return err
			}
			pos.Shares += in.Amount
			receiving = append(receiving, pos)
		}

		cash := (k - 1) * in.Amount * shareValue
		carried, realized := released-cash, int64(0)
		if carried < 0 || len(receiving) == 0 {
			carried, realized = 0, cash-released
		}
		for i, pos := range receiving {
			share := carried / int64(len(receiving))
			if i == 0 {
				share += carried % int64(len(receiving))
			}
			pos.CostBasis += share
		}
		sess.addPnL(in.UserID, 0, model.PnLConversion, realized)

		if cash > 0 {
			bal, err := sess.balance(ctx, in.UserID)
			if err != nil {
				return err
			}
			bal.Available += cash
			if err := s.ledger.Create(ctx, &model.LedgerEntry{UserID: in.UserID, Type: model.LedgerConversion, Amount: cash}); err != nil {
				return err
			}
		}
		if err := sess.flush(ctx); err != nil {
			return err
		}

		result = &port.ConvertResult{Cash: cash}
		for _, pos := range receiving {
			result.YesMarkets = append(result.YesMarkets, pos.MarketID)
		}
		return nil
	})
	return result, err
}
