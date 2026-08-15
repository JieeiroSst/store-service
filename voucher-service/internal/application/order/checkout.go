package order

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

func (s *Service) Checkout(ctx context.Context, in CheckoutInput) (*CheckoutOutput, error) {
	var out *CheckoutOutput

	err := s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		o, err := s.repo.FindByIDForUpdate(ctx, in.OrderID)
		if err != nil {
			return err
		}
		if o.Status != order.StatusAwaitingPayment {
			return order.ErrInvalidOrderTransition
		}

		if in.PaymentMethod == "wallet" {
			ownerType := "user"
			if o.BuyerType == order.BuyerTypeCorporate {
				ownerType = "corporate"
			}
			if err := s.walletDebiter.Debit(ctx, ownerType, o.BuyerID, o.TotalAmount, "order:"+o.ID.String()); err != nil {
				return err
			}
			if err := s.fulfill(ctx, o, "wallet:"+o.ID.String()); err != nil {
				return err
			}
			if err := s.repo.Save(ctx, o); err != nil {
				return err
			}
			if err := s.enqueueEvents(ctx, o.PullEvents()); err != nil {
				return err
			}
			out = &CheckoutOutput{OrderID: o.ID.String(), Status: string(o.Status)}
			return nil
		}

		intent, err := s.paymentInitiator.InitiatePayment(ctx, PaymentRequest{
			OrderID: o.ID,
			Amount:  o.TotalAmount,
			Method:  in.PaymentMethod,
		})
		if err != nil {
			return err
		}
		out = &CheckoutOutput{OrderID: o.ID.String(), Status: string(o.Status), PaymentAction: &intent}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) ConfirmPayment(ctx context.Context, id shared.OrderID, paymentRef string) error {
	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		o, err := s.repo.FindByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if err := o.MarkPaid(paymentRef, s.clock.Now()); err != nil {
			return err
		}
		if err := s.fulfill(ctx, o, paymentRef); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, o); err != nil {
			return err
		}
		return s.enqueueEvents(ctx, o.PullEvents())
	})
}

func (s *Service) fulfill(ctx context.Context, o *order.Order, paymentRef string) error {
	if o.Status == order.StatusAwaitingPayment {
		if err := o.MarkPaid(paymentRef, s.clock.Now()); err != nil {
			return err
		}
	}
	if err := o.MarkFulfilling(s.clock.Now()); err != nil {
		return err
	}

	items := make([]VoucherIssuanceItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, VoucherIssuanceItem{
			MerchantID:   item.MerchantID,
			ProductSKU:   item.ProductSKU,
			Denomination: item.UnitPrice,
			Quantity:     item.Quantity,
		})
	}
	refs, err := s.voucherIssuer.IssueVouchersForOrder(ctx, VoucherIssuanceRequest{OrderID: o.ID, Items: items})
	if err != nil {
		return err
	}

	cursor := 0
	for i, item := range o.Items {
		voucherIDs := make([]shared.VoucherID, 0, item.Quantity)
		for j := 0; j < item.Quantity && cursor < len(refs); j++ {
			id, err := shared.ParseVoucherID(refs[cursor].VoucherID)
			if err != nil {
				return err
			}
			voucherIDs = append(voucherIDs, id)
			cursor++
		}
		if err := o.AttachIssuedVouchers(i, voucherIDs); err != nil {
			return err
		}
	}

	return o.Complete(s.clock.Now())
}

func (s *Service) CancelOrder(ctx context.Context, id shared.OrderID, reason string) error {
	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		o, err := s.repo.FindByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if err := o.Cancel(reason, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, o); err != nil {
			return err
		}
		return s.enqueueEvents(ctx, o.PullEvents())
	})
}
