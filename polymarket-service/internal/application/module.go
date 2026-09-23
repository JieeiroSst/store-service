package application

import (
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

type Options struct {
	TreasuryWalletID string
	Currency         string
	ShareValue       int64
	MinOrderSize     int64
	DisputeWindow    time.Duration
	DisputeBond      int64
	TakerFeeBps      int64
	MakerRebateBps   int64
	ReferralBps      int64
	RewardEpoch      time.Duration
}

func (o Options) withDefaults() Options {
	if o.Currency == "" {
		o.Currency = "USD"
	}
	if o.ShareValue < 2 {
		o.ShareValue = 100
	}
	if o.MinOrderSize < 1 {
		o.MinOrderSize = 1
	}
	if o.DisputeWindow <= 0 {
		o.DisputeWindow = 2 * time.Hour
	}
	if o.RewardEpoch <= 0 {
		o.RewardEpoch = time.Minute
	}
	o.TakerFeeBps = clampBps(o.TakerFeeBps)
	o.MakerRebateBps = clampBps(o.MakerRebateBps)
	o.ReferralBps = clampBps(o.ReferralBps)
	if o.MakerRebateBps+o.ReferralBps > 10_000 {
		o.ReferralBps = 10_000 - o.MakerRebateBps
	}
	return o
}

func clampBps(v int64) int64 { return min(max(v, 0), 10_000) }

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewCatalogService, fx.As(new(port.CatalogUsecase))),
		fx.Annotate(NewResolutionService, fx.As(new(port.ResolutionUsecase))),
		fx.Annotate(NewExchangeService, fx.As(new(port.ExchangeUsecase))),
		fx.Annotate(NewAccountService, fx.As(new(port.AccountUsecase))),
		fx.Annotate(NewSocialService, fx.As(new(port.SocialUsecase))),
		fx.Annotate(NewRewardService, fx.As(new(port.RewardUsecase))),
		fx.Annotate(NewReferralService, fx.As(new(port.ReferralUsecase))),
		fx.Annotate(NewIdentityService, fx.As(new(port.IdentityUsecase))),
	),
)
