package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/JIeeiroSst/card-service/config"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/postgres"
	"github.com/JIeeiroSst/card-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(New),
)

type Ports struct {
	fx.Out

	Accounts       port.AccountRepository
	Cards          port.CardRepository
	Authorizations port.AuthorizationRepository
	Transactions   port.TransactionRepository
	Locker         port.Locker
	Health         port.HealthChecker
}

func New(lc fx.Lifecycle, cfg *config.Config) (Ports, error) {
	switch cfg.Storage.Backend {
	case "memory":
		log.Printf("storage backend: memory (data is lost on restart)")
		s := memory.NewStore()
		return Ports{
			Accounts:       memory.NewAccounts(s),
			Cards:          memory.NewCards(s),
			Authorizations: memory.NewAuthorizations(s),
			Transactions:   memory.NewTransactions(s),
			Locker:         memory.NewLocker(),
			Health:         noop{},
		}, nil
	case "postgres":
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Postgres.ConnectTimeout)
		defer cancel()
		db, err := postgres.Connect(ctx, cfg.Postgres.DSN(), cfg.Postgres.MaxConns)
		if err != nil {
			return Ports{}, err
		}
		if err := db.Migrate(ctx); err != nil {
			db.Close()
			return Ports{}, fmt.Errorf("migrate postgres: %w", err)
		}
		lc.Append(fx.Hook{
			OnStop: func(context.Context) error {
				db.Close()
				return nil
			},
		})
		log.Printf("storage backend: postgres %s/%s", cfg.Postgres.Host, cfg.Postgres.Database)
		return Ports{
			Accounts:       postgres.NewAccounts(db),
			Cards:          postgres.NewCards(db),
			Authorizations: postgres.NewAuthorizations(db),
			Transactions:   postgres.NewTransactions(db),
			Locker:         db,
			Health:         db,
		}, nil
	}
	return Ports{}, fmt.Errorf("unknown STORAGE_BACKEND %q (memory or postgres)", cfg.Storage.Backend)
}

type noop struct{}

func (noop) Ping(context.Context) error { return nil }
