package mysql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	pkgmysql "github.com/referral/service/pkg/mysql"
)

var NewClient = pkgmysql.NewClient

var Module = fx.Options(
	fx.Provide(
		NewClient,
		NewReferralLinkRepo,
		NewReferralEventRepo,
		NewRewardRepo,
		NewUserStatsRepo,
		NewRewardProgramRepo,
		NewAttributionRepo,
	),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, db *sqlx.DB) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return db.Close()
		},
	})
}
