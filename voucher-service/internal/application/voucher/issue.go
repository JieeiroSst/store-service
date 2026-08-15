package voucher

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"go.uber.org/zap"
)

func (s *Service) IssueVouchers(ctx context.Context, in IssueVouchersInput) ([]IssuedVoucher, error) {
	if in.Quantity <= 0 {
		return nil, shared.ErrValidation
	}

	info, err := s.merchantLookup.GetMerchantInfo(ctx, in.MerchantID)
	if err != nil {
		return nil, err
	}
	if !info.Active {
		return nil, merchant.ErrMerchantInactive
	}

	provider, err := s.registry.Resolve(info.ProviderType)
	if err != nil {
		return nil, err
	}

	ref := shared.ProductRef{
		MerchantID:   in.MerchantID,
		SKU:          in.ProductSKU,
		Denomination: in.Denomination,
	}

	codes, err := provider.Issue(ctx, ref, in.Quantity)
	if err != nil {
		return nil, fmt.Errorf("provider issue: %w", err)
	}
	if len(codes) != in.Quantity {
		return nil, fmt.Errorf("provider returned %d codes, wanted %d", len(codes), in.Quantity)
	}

	var expiresAt *time.Time
	if in.ExpiresInDays > 0 {
		t := s.clock.Now().AddDate(0, 0, in.ExpiresInDays)
		expiresAt = &t
	}

	issued := make([]IssuedVoucher, 0, in.Quantity)
	err = s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		for _, code := range codes {
			now := s.clock.Now()
			v, err := voucher.NewVoucher(in.MerchantID, ref, now)
			if err != nil {
				return err
			}
			v.OrderID = in.OrderID
			if err := v.Issue(code, expiresAt, now); err != nil {
				return err
			}
			if err := s.repo.Create(ctx, v); err != nil {
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
			issued = append(issued, IssuedVoucher{Voucher: v, PlaintextPIN: code.PIN})
		}
		return nil
	})
	if err != nil {
		s.log.Error("issue vouchers failed", zap.Error(err))
		return nil, err
	}

	return issued, nil
}
