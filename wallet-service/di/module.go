package di

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/adapter/primary/httpapi"
	"github.com/JIeeiroSst/wallet-service/internal/adapter/secondary/bankgateway"
	"github.com/JIeeiroSst/wallet-service/internal/adapter/secondary/pinhash"
	"github.com/JIeeiroSst/wallet-service/internal/adapter/secondary/postgres"
	"github.com/JIeeiroSst/wallet-service/internal/adapter/secondary/rediscache"
	"github.com/JIeeiroSst/wallet-service/internal/application"
	"github.com/JIeeiroSst/wallet-service/internal/config"
	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
	"github.com/JIeeiroSst/wallet-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/wallet-service/internal/infrastructure/logger"
	"github.com/JIeeiroSst/wallet-service/internal/infrastructure/server"
)

var InfrastructureModule = fx.Module("infrastructure",
	fx.Provide(
		config.Load,
		func(cfg *config.Config) (*zap.Logger, error) {
			log, err := logger.New(cfg.Log.Level)
			if err != nil {
				return nil, err
			}
			log.Info("configuration loaded", zap.String("source", cfg.Source))
			return log, nil
		},
		database.NewPostgres,
		database.NewRedis,
	),
)

var AdapterModule = fx.Module("adapters",
	fx.Provide(
		postgres.NewTxManager,
		postgres.NewWalletRepository,
		postgres.NewTransactionRepository,
		postgres.NewLedgerRepository,
		postgres.NewCardRepository,
		postgres.NewBankRepository,
		postgres.NewMerchantRepository,
		postgres.NewBankAccountRepository,
		postgres.NewPaymentRequestRepository,
		postgres.NewSettlementBatchRepository,
		postgres.NewClearingRepository,
		func(client *redis.Client) port.Cache { return rediscache.New(client) },
		func(cfg *config.Config) port.PinHasher { return pinhash.NewBcrypt(cfg.Security.PinBcryptCost) },
		func(cfg *config.Config, log *zap.Logger) (port.BankGateway, error) {
			limit, err := decimal.NewFromString(cfg.BankGateway.DeclineAbove)
			if err != nil {
				return nil, err
			}
			return bankgateway.NewSimulated(limit, log), nil
		},
	),
)

type depsIn struct {
	fx.In

	Tx              port.TxManager
	Wallets         port.WalletRepository
	Transactions    port.TransactionRepository
	Ledger          port.LedgerRepository
	Cards           port.CardRepository
	Banks           port.BankRepository
	Merchants       port.MerchantRepository
	BankAccounts    port.BankAccountRepository
	PaymentRequests port.PaymentRequestRepository
	Batches         port.SettlementBatchRepository
	Clearing        port.ClearingRepository
	Cache           port.Cache
	Pins            port.PinHasher
	BankGateway     port.BankGateway
	Config          *config.Config
	Logger          *zap.Logger
}

func provideDeps(in depsIn) (application.Deps, error) {
	fees, err := feeSchedule(in.Config.Fees)
	if err != nil {
		return application.Deps{}, err
	}
	return application.Deps{
		Tx:              in.Tx,
		Wallets:         in.Wallets,
		Transactions:    in.Transactions,
		Ledger:          in.Ledger,
		Cards:           in.Cards,
		Banks:           in.Banks,
		Merchants:       in.Merchants,
		BankAccounts:    in.BankAccounts,
		PaymentRequests: in.PaymentRequests,
		Batches:         in.Batches,
		Clearing:        in.Clearing,
		Cache:           in.Cache,
		Pins:            in.Pins,
		CardHashKey:     []byte(in.Config.Security.CardHashKey),
		BankGateway:     in.BankGateway,
		Fees:            fees,
		PaymentRequest:  application.PaymentRequestPolicy{TTL: in.Config.PaymentRequests.TTL()},
		Clock:           time.Now,
		Logger:          in.Logger,
	}, nil
}

func feeSchedule(cfg config.FeesConfig) (model.FeeSchedule, error) {
	fees := model.DefaultFeeSchedule()
	for _, f := range []struct {
		raw string
		dst *decimal.Decimal
	}{
		{cfg.WithdrawalRate, &fees.WithdrawalRate},
		{cfg.WithdrawalMinUSD, &fees.WithdrawalMinUSD},
		{cfg.TransferRate, &fees.TransferRate},
	} {
		if f.raw == "" {
			continue
		}
		v, err := decimal.NewFromString(f.raw)
		if err != nil {
			return model.FeeSchedule{}, err
		}
		*f.dst = v
	}
	return fees, nil
}

var ApplicationModule = fx.Module("application",
	fx.Provide(
		provideDeps,
		application.NewWalletService,
		application.NewFundingService,
		application.NewTransferService,
		application.NewPaymentService,
		application.NewCardPaymentService,
		application.NewSettlementService,
		application.NewHistoryService,
		application.NewCardService,
		application.NewBankAccountService,
		application.NewPaymentRequestService,
	),
)

type servicesIn struct {
	fx.In

	Wallets         port.WalletService
	Funding         port.FundingService
	Transfers       port.TransferService
	Payments        port.PaymentService
	CardPayments    port.CardPaymentService
	Settlements     port.SettlementService
	History         port.HistoryService
	Cards           port.CardService
	BankAccounts    port.BankAccountService
	PaymentRequests port.PaymentRequestService
}

func provideRouter(in servicesIn, db *sqlx.DB, cfg *config.Config, log *zap.Logger) *gin.Engine {
	return httpapi.NewRouter(httpapi.Services{
		Wallets:         in.Wallets,
		Funding:         in.Funding,
		Transfers:       in.Transfers,
		Payments:        in.Payments,
		CardPayments:    in.CardPayments,
		Settlements:     in.Settlements,
		History:         in.History,
		Cards:           in.Cards,
		BankAccounts:    in.BankAccounts,
		PaymentRequests: in.PaymentRequests,
	}, httpapi.RouterOptions{
		RateLimitCapacity: cfg.Server.RateLimitCapacity,
		RateLimitPerSec:   cfg.Server.RateLimitPerSec,
		Ready:             db.PingContext,
	}, log)
}

var HTTPModule = fx.Module("http",
	fx.Provide(
		provideRouter,
		func(engine *gin.Engine, cfg *config.Config) *http.Server { return server.New(engine, cfg) },
	),
	fx.Invoke(server.Register),
)

var App = fx.Options(
	InfrastructureModule,
	AdapterModule,
	ApplicationModule,
	HTTPModule,
)
