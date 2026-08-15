package voucher

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
)

func (s *Service) RevokeVoucher(ctx context.Context, id shared.VoucherID, reason string) error {
	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		v, err := s.repo.FindByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if err := v.Revoke(reason, s.clock.Now()); err != nil {
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
		return nil
	})
}

func (s *Service) GetVoucher(ctx context.Context, id shared.VoucherID) (*voucher.Voucher, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListVouchers(ctx context.Context, ownerType voucher.OwnerType, ownerID string) ([]*voucher.Voucher, error) {
	return s.repo.ListByOwner(ctx, ownerType, ownerID)
}
