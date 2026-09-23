package infrastructure

import (
	"errors"

	"github.com/JIeeiroSst/polymarket-service/config"
	httpadapter "github.com/JIeeiroSst/polymarket-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/notification"
	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/realtime"
	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/referral"
	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/polymarket-service/internal/adapter/secondary/wallet"
	"github.com/JIeeiroSst/polymarket-service/internal/application"
	"github.com/JIeeiroSst/polymarket-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/polymarket-service/internal/infrastructure/jobs"
	"github.com/JIeeiroSst/polymarket-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

func newApplicationOptions(cfg *config.Config) (application.Options, error) {
	if cfg.Wallet.TreasuryWalletID == "" {
		return application.Options{}, errors.New("treasury wallet id is required (TREASURY_WALLET_ID or wallet.treasuryWalletID in Consul)")
	}
	return application.Options{
		TreasuryWalletID: cfg.Wallet.TreasuryWalletID,
		Currency:         cfg.Exchange.Currency,
		ShareValue:       cfg.Exchange.ShareValue,
		MinOrderSize:     cfg.Exchange.MinOrderSize,
		DisputeWindow:    cfg.Exchange.DisputeWindowDuration(),
		DisputeBond:      cfg.Exchange.DisputeBond,
		TakerFeeBps:      cfg.Fees.TakerFeeBps,
		MakerRebateBps:   cfg.Fees.MakerRebateBps,
		ReferralBps:      cfg.Fees.ReferralBps,
		RewardEpoch:      cfg.Exchange.RewardEpochDuration(),
	}, nil
}

var Module = fx.Options(
	fx.Provide(newConfig, newApplicationOptions),

	database.Module, // *gorm.DB

	repository.Module,   // event, market, order, trade, balance, ledger, position, social, stats repositories + TxManager
	realtime.Module,     // port.Publisher + port.EventStream (SSE hub)
	wallet.Module,       // port.WalletGateway -> payment-wallet-service
	userservice.Module,  // port.UserDirectory -> user-service (bearer token validation)
	referral.Module,     // port.ReferralGateway -> referral-service
	notification.Module, // port.Notifier -> notification-service

	application.Module, // catalog, resolution, exchange, account and social use cases

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New, jobs.RegisterRewardJob),
)
