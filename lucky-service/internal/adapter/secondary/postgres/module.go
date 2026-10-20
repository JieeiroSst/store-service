package postgres

import (
	"context"

	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewRepository,
		NewPlayRepository,
		NewFraudRepository,
		fx.Annotate(func(r *FraudRepository) *FraudRepository { return r }, fx.As(new(port.FraudRepository))),
		fx.Annotate(func(r *Repository) *Repository { return r }, fx.As(new(port.CampaignRepository))),
		fx.Annotate(func(r *Repository) *Repository { return r }, fx.As(new(port.HealthChecker)), fx.ResultTags(`group:"health"`)),
		fx.Annotate(func(r *PlayRepository) *PlayRepository { return r }, fx.As(new(port.PlayRepository))),
		fx.Annotate(func(r *PlayRepository) *PlayRepository { return r }, fx.As(new(port.RewardOutbox))),
	),
	fx.Invoke(registerMigration),
)

func registerMigration(lc fx.Lifecycle, r *Repository) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return r.Migrate(ctx) },
	})
}
