package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type PaymentRail struct {
	wallets outbound.WalletGateway
	gateway outbound.PaymentGateway
	log     *slog.Logger
}

func NewPaymentRail(w outbound.WalletGateway, g outbound.PaymentGateway, log *slog.Logger) *PaymentRail {
	return &PaymentRail{wallets: w, gateway: g, log: defaultLog(log)}
}

func (r *PaymentRail) chargeWallet(ctx context.Context, userID int64, toWalletID, currency string, amount int64, reference, desc string) (string, error) {
	if toWalletID == "" {
		return "", invalid("this event does not accept wallet payments")
	}
	if r.wallets == nil {
		return "", invalid("wallet payment is not enabled")
	}
	w, err := r.wallets.GetByUser(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return "", invalid("you have no wallet yet, create one first")
	}
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(w.Currency, currency) {
		return "", invalid("wallet currency %s does not match %s", w.Currency, currency)
	}
	return r.wallets.Transfer(ctx, outbound.TransferParams{
		FromWalletID: w.ID, ToWalletID: toWalletID, Amount: amount, ReferenceID: reference, Description: desc,
	})
}

func (r *PaymentRail) undoWallet(ctx context.Context, transferID, reason string) {
	if err := r.wallets.ReverseTransfer(ctx, transferID, reason); err != nil {
		r.log.Error("refund of undeliverable charge failed", "transfer", transferID, "err", err)
	}
}

func (r *PaymentRail) gatewayPay(ctx context.Context, existingRef string, in outbound.CreateGatewayPayment) (p domain.GatewayPayment, fresh bool, err error) {
	if r.gateway == nil {
		return p, false, invalid("gateway payment is not enabled")
	}
	if in.Provider == "" {
		return p, false, invalid("provider is required for gateway payment")
	}
	if id, perr := strconv.ParseInt(existingRef, 10, 64); perr == nil {
		p, err = r.gateway.Get(ctx, id)
		if err != nil {
			return p, false, err
		}
		if p.Status != domain.GatewayFailed && p.Status != domain.GatewayRefunded {
			return p, false, nil
		}
	}
	p, err = r.gateway.Create(ctx, in)
	return p, true, err
}

func (r *PaymentRail) gatewayStatus(ctx context.Context, ref string) (domain.GatewayPayment, bool) {
	if r.gateway == nil {
		return domain.GatewayPayment{}, false
	}
	id, err := strconv.ParseInt(ref, 10, 64)
	if err != nil {
		return domain.GatewayPayment{}, false
	}
	p, err := r.gateway.Get(ctx, id)
	return p, err == nil
}

func (r *PaymentRail) refund(ctx context.Context, method domain.PaymentMethod, ref, reason string, notCaptured bool) error {
	if ref == "" || method == domain.MethodFree {
		return nil
	}
	switch method {
	case domain.MethodWallet:
		if r.wallets == nil {
			return fmt.Errorf("%w: wallet payment is not enabled, cannot refund", domain.ErrUpstreamUnavailable)
		}
		return r.wallets.ReverseTransfer(ctx, ref, reason)
	case domain.MethodGateway:
		if r.gateway == nil {
			return fmt.Errorf("%w: gateway payment is not enabled, cannot refund", domain.ErrUpstreamUnavailable)
		}
		id, err := strconv.ParseInt(ref, 10, 64)
		if err != nil {
			return nil
		}
		err = r.gateway.Refund(ctx, id)
		if errors.Is(err, domain.ErrConflict) && notCaptured {
			return nil
		}
		return err
	}
	return nil
}

type staleAction int

const (
	staleKeep    staleAction = iota 
	stalePaid                      
	staleRelease                    
)

func (r *PaymentRail) judgeStale(ctx context.Context, ref string, expiresAt time.Time, grace time.Duration, now time.Time) staleAction {
	p, ok := r.gatewayStatus(ctx, ref)
	if !ok {
		return staleKeep
	}
	if p.Status == domain.GatewayCaptured {
		return stalePaid
	}
	inFlight := p.Status == domain.GatewayPending || p.Status == domain.GatewayAuthorized
	if inFlight && now.Before(expiresAt.Add(grace)) {
		return staleKeep
	}
	if p.Status == domain.GatewayAuthorized {
		if err := r.gateway.Refund(ctx, p.ID); err != nil {
			r.log.Error("release authorization", "payment", p.ID, "err", err)
			return staleKeep
		}
	}
	return staleRelease
}
