package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

func TestCheckoutAppliesCouponDiscount(t *testing.T) {
	f := newFixture(t)
	f.coupons.discount = 50
	_, r := f.reserve(t)
	out, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentCard, CouponCode: " SAVE50 "})
	if err != nil {
		t.Fatal(err)
	}
	if out.Payment.AmountCents != 100 || out.Payment.DiscountCents != 50 || out.Payment.CouponCode != "SAVE50" {
		t.Fatalf("payment %+v", out.Payment)
	}
	if f.gateway.last.AmountCents != 100 {
		t.Fatalf("charged %d", f.gateway.last.AmountCents)
	}
	if len(f.coupons.redeemed) != 1 || f.coupons.redeemed[0] != out.Order.OrderNo {
		t.Fatalf("redeemed %v, order_no %d", f.coupons.redeemed, out.Order.OrderNo)
	}
}

func TestCheckoutFreeItemSkipsGateway(t *testing.T) {
	f := newFixture(t)
	f.coupons.discount = 500
	_, r := f.reserve(t)
	out, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentCash, CouponCode: "FREE"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Payment.AmountCents != 0 || out.Payment.DiscountCents != 150 || f.gateway.charges != 0 {
		t.Fatalf("payment %+v charges %d", out.Payment, f.gateway.charges)
	}
	if _, err := f.svc.RefundOrder(context.Background(), out.Order.ID, "complaint"); err == nil {
		t.Fatal("processing order must not be refundable")
	}
	f.svc.ReportDispense(context.Background(), out.Order.ID, true)
	if _, err := f.svc.RefundOrder(context.Background(), out.Order.ID, "complaint"); err != nil || len(f.gateway.refunds) != 0 {
		t.Fatalf("err=%v refunds=%v", err, f.gateway.refunds)
	}
}

func TestCheckoutRejectedCouponChargesNothing(t *testing.T) {
	f := newFixture(t)
	f.coupons.quoteErr = domain.ErrCouponRejected
	_, r := f.reserve(t)
	_, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentCard, CouponCode: "OLD"})
	if !errors.Is(err, domain.ErrCouponRejected) || f.gateway.charges != 0 || len(f.st.payments) != 0 {
		t.Fatalf("err=%v charges=%d payments=%d", err, f.gateway.charges, len(f.st.payments))
	}
}

func TestCouponRedeemFailureStillSells(t *testing.T) {
	f := newFixture(t)
	f.coupons.discount = 20
	f.coupons.redeemErr = domain.ErrUpstream
	_, r := f.reserve(t)
	if _, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentCard, CouponCode: "X"}); err != nil {
		t.Fatal(err)
	}
	if !f.hasEvent(domain.EventCouponRedeemFailed) {
		t.Fatal("expected coupon.redeem_failed alert")
	}
}

func TestCheckoutWalletNeedsWalletID(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	if _, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentWallet}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
	if _, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentCard, WalletID: "w1"}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
	out, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentWallet, WalletID: "w1"})
	if err != nil {
		t.Fatal(err)
	}
	if f.gateway.last.WalletID != "w1" || f.gateway.last.PaymentID != out.Payment.ID || out.Payment.Metadata["wallet_id"] != "w1" {
		t.Fatalf("charge %+v metadata %v", f.gateway.last, out.Payment.Metadata)
	}
}

func TestCheckoutRejectsPriceChangeDuringQuote(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	f.coupons.discount = 10
	f.svc.coupons = quoteHook{f.coupons, func() {
		p := f.st.products["p1"]
		p.PriceCents = 999
		f.st.products["p1"] = p
	}}
	_, err := f.svc.Checkout(context.Background(), r.ID, port.CheckoutInput{Method: domain.PaymentCard, CouponCode: "X"})
	if !errors.Is(err, domain.ErrPriceChanged) || f.gateway.charges != 0 {
		t.Fatalf("err=%v charges=%d", err, f.gateway.charges)
	}
}

type quoteHook struct {
	*fakeCoupons
	during func()
}

func (q quoteHook) Quote(ctx context.Context, code string, amount int) (int, error) {
	q.during()
	return q.fakeCoupons.Quote(ctx, code, amount)
}

func TestRefundOrder(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	out, _ := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard))
	f.svc.ReportDispense(context.Background(), out.Order.ID, true)

	if _, err := f.svc.RefundOrder(context.Background(), out.Order.ID, " "); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
	o, err := f.svc.RefundOrder(context.Background(), out.Order.ID, "stale product")
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != domain.OrderRefunded || f.st.payments[out.Payment.ID].Status != domain.PaymentRefunded {
		t.Fatalf("order=%s payment=%s", o.Status, f.st.payments[out.Payment.ID].Status)
	}
	if _, err := f.svc.RefundOrder(context.Background(), out.Order.ID, "again"); err != nil || len(f.gateway.refunds) != 1 {
		t.Fatalf("second refund: err=%v refunds=%v", err, f.gateway.refunds)
	}
}

func TestRefundFailureRaisesAlert(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	out, _ := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard))
	f.gateway.refundErr = errors.New("processor down")
	if _, err := f.svc.ReportDispense(context.Background(), out.Order.ID, false); !errors.Is(err, domain.ErrPaymentFailed) {
		t.Fatalf("got %v", err)
	}
	if !f.hasEvent(domain.EventRefundFailed) || !f.hasEvent(domain.EventOrderFailed) {
		t.Fatalf("events %v", f.st.events)
	}
}

func TestReserveLastUnitRaisesEmptyAlert(t *testing.T) {
	f := newFixture(t)
	inv := f.st.inventory["i1"]
	inv.Quantity = 1
	f.st.inventory["i1"] = inv
	f.reserve(t)
	if !f.hasEvent(domain.EventInventoryEmpty) || f.hasEvent(domain.EventInventoryLow) {
		t.Fatalf("events %v", f.st.events)
	}
}

func TestListSessionOrders(t *testing.T) {
	f := newFixture(t)
	sess, r := f.reserve(t)
	f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCash))
	orders, err := f.svc.ListSessionOrders(context.Background(), sess.ID)
	if err != nil || len(orders) != 1 {
		t.Fatalf("err=%v orders=%d", err, len(orders))
	}
	if _, err := f.svc.ListSessionOrders(context.Background(), "missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}
