package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	ticketservice "github.com/JIeeiroSst/ticket-service"
	"github.com/JIeeiroSst/ticket-service/config"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const schemaLockID = 727303

func OpenPool(ctx context.Context, dsn string, maxConns int) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres config: %w", err)
	}
	pcfg.MaxConns = int32(maxConns)
	pcfg.MinConns = min(int32(maxConns), 4)
	pcfg.MaxConnLifetime = 30 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return pool, nil
}

func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	pool, err := OpenPool(ctx, cfg.PostgresDSN(), cfg.PostgresMaxConns)
	if err != nil {
		return nil, err
	}
	if err := ApplySchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func ApplySchema(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "select pg_advisory_lock($1)", schemaLockID); err != nil {
		return err
	}
	defer conn.Exec(context.WithoutCancel(ctx), "select pg_advisory_unlock($1)", schemaLockID)
	if _, err := conn.Exec(ctx, ticketservice.Schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505", "23514":
			return domain.ErrConflict
		case "23503":
			return fmt.Errorf("%w: referenced record does not exist", domain.ErrInvalid)
		case "55P03", "57014", "40001", "40P01", "53300", "57P03":
			return fmt.Errorf("%w: database is busy", domain.ErrBusy)
		}
	}
	return err
}
