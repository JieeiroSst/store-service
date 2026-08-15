package voucher

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"go.uber.org/zap"
)

func (s *Service) ExpireDueVouchers(ctx context.Context) (int, error) {
	due, err := s.repo.ListDueForExpiry(ctx, s.clock.Now())
	if err != nil {
		return 0, err
	}

	expired := 0
	for _, v := range due {
		err := s.txManager.WithinTx(ctx, func(ctx context.Context) error {
			fresh, err := s.repo.FindByIDForUpdate(ctx, v.ID)
			if err != nil {
				return err
			}
			if !fresh.IsExpired(s.clock.Now()) {
				return nil
			}
			if err := fresh.Expire(s.clock.Now()); err != nil {
				return err
			}
			if err := s.repo.Save(ctx, fresh); err != nil {
				return err
			}
			for _, evt := range fresh.PullEvents() {
				outboxEvt, err := outbox.NewEventFromDomain(aggregateType, outboxTopic, evt)
				if err != nil {
					return err
				}
				if err := s.outboxP.Enqueue(ctx, outboxEvt); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			s.log.Error("expire voucher failed", zap.String("voucher_id", v.ID.String()), zap.Error(err))
			continue
		}
		expired++
	}
	return expired, nil
}
