package internalgateway

import (
	"context"

	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	walletapp "github.com/JIeeiroSst/voucher-service/internal/application/wallet"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/wallet"
)

type WalletDebiterAdapter struct {
	walletSvc walletapp.WalletService
}

func NewWalletDebiterAdapter(walletSvc walletapp.WalletService) orderapp.WalletDebiter {
	return &WalletDebiterAdapter{walletSvc: walletSvc}
}

func (a *WalletDebiterAdapter) Debit(ctx context.Context, ownerType, ownerID string, amount shared.Money, reason string) error {
	return a.walletSvc.Debit(ctx, wallet.OwnerType(ownerType), ownerID, amount, "order", reason, "")
}
