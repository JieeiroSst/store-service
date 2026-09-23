package repository

import (
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewEventRepository, fx.As(new(port.EventRepository))),
		fx.Annotate(NewMarketRepository, fx.As(new(port.MarketRepository))),
		fx.Annotate(NewOrderRepository, fx.As(new(port.OrderRepository))),
		fx.Annotate(NewTradeRepository, fx.As(new(port.TradeRepository))),
		fx.Annotate(NewBalanceRepository, fx.As(new(port.BalanceRepository))),
		fx.Annotate(NewLedgerRepository, fx.As(new(port.LedgerRepository))),
		fx.Annotate(NewPositionRepository, fx.As(new(port.PositionRepository))),
		fx.Annotate(NewSocialRepository, fx.As(new(port.SocialRepository))),
		fx.Annotate(NewStatsRepository, fx.As(new(port.StatsRepository))),
		fx.Annotate(NewPnLRepository, fx.As(new(port.PnLRepository))),
		fx.Annotate(NewExchangeAccountRepository, fx.As(new(port.ExchangeAccountRepository))),
		fx.Annotate(NewRewardRepository, fx.As(new(port.RewardRepository))),
		fx.Annotate(NewReferralRepository, fx.As(new(port.ReferralRepository))),
		fx.Annotate(NewTxManager, fx.As(new(port.TxManager))),
	),
)
