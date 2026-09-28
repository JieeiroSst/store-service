package payment

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

var errWalletDisabled = errors.New("wallet payments are not configured")

type Router struct {
	wallet   port.PaymentGateway
	terminal port.PaymentGateway
}

func NewRouter(wallet, terminal port.PaymentGateway) *Router {
	return &Router{wallet: wallet, terminal: terminal}
}

func (r *Router) pick(method domain.PaymentMethod) (port.PaymentGateway, error) {
	if method != domain.PaymentWallet {
		return r.terminal, nil
	}
	if r.wallet == nil {
		return nil, errWalletDisabled
	}
	return r.wallet, nil
}

func (r *Router) Charge(ctx context.Context, req port.ChargeRequest) (string, error) {
	g, err := r.pick(req.Method)
	if err != nil {
		return "", err
	}
	return g.Charge(ctx, req)
}

func (r *Router) Lookup(ctx context.Context, req port.LookupRequest) (string, bool, error) {
	g, err := r.pick(req.Method)
	if err != nil {
		return "", false, err
	}
	return g.Lookup(ctx, req)
}

func (r *Router) Refund(ctx context.Context, req port.RefundRequest) error {
	g, err := r.pick(req.Method)
	if err != nil {
		return err
	}
	return g.Refund(ctx, req)
}
