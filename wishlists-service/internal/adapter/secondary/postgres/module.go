package postgres

import (
	"context"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewRepository,
		fx.Annotate(func(r *Repository) *Repository { return r }, fx.As(new(port.WishlistRepository))),
		fx.Annotate(func(r *Repository) *Repository { return r }, fx.As(new(port.HealthChecker)), fx.ResultTags(`group:"health"`)),
	),
	fx.Invoke(registerMigration),
)

func registerMigration(lc fx.Lifecycle, r *Repository) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return r.Migrate(ctx) },
	})
}
