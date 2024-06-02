package timescale

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
)

var (
	instance *TimescaleConnect
	once     sync.Once
)

type TimescaleConnect struct {
	db *pgx.Conn
}

func GetMysqlConnInstance(ctx context.Context, dns string) *TimescaleConnect {
	once.Do(func() {
		db, err := pgx.Connect(ctx, dns)
		if err != nil {
			return
		}
		stmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v;", "billing")
		db.Exec(ctx, stmt)

		instance = &TimescaleConnect{db: db}
	})
	return instance
}

func NewTimescaleConn(ctx context.Context, dns string) *pgx.Conn {
	return GetMysqlConnInstance(ctx, dns).db
}
