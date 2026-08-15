package voucher

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
)

func (s *Service) ActivateVoucher(ctx context.Context, id shared.VoucherID, ownerType voucher.OwnerType, ownerID string) (*voucher.Voucher, error) {
	var activated *voucher.Voucher
	err := s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		v, err := s.repo.FindByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if err := v.Activate(ownerType, ownerID, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, v); err != nil {
			return err
		}
		for _, evt := range v.PullEvents() {
			outboxEvt, err := outbox.NewEventFromDomain(aggregateType, outboxTopic, evt)
			if err != nil {
				return err
			}
			if err := s.outboxP.Enqueue(ctx, outboxEvt); err != nil {
				return err
			}
		}
		activated = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return activated, nil
}
