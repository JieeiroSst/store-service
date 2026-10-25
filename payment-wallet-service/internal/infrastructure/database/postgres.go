package database

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/config"
	"github.com/JIeeiroSst/utils/postgres"
	"go.uber.org/fx"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	db := openPostgres(postgres.PostgresConfig{
		PostgresqlHost:     cfg.Postgres.PostgresqlHost,
		PostgresqlPort:     cfg.Postgres.PostgresqlPort,
		PostgresqlUser:     cfg.Postgres.PostgresqlUser,
		PostgresqlPassword: cfg.Postgres.PostgresqlPassword,
		PostgresqlDbname:   cfg.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  cfg.Postgres.PostgresqlSSLMode,
	})
	if err := applySchema(db, "database.sql"); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return db, nil
}

func applySchema(db *gorm.DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, stmt := range splitStatements(string(raw)) {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("exec statement %q: %w", stmt, err)
		}
	}
	return nil
}

// splitStatements splits a SQL script on semicolons after dropping full-line "--" comments, so a
// semicolon inside a comment cannot cut a statement in two.
func splitStatements(script string) []string {
	var kept []string
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		kept = append(kept, line)
	}
	var out []string
	for _, stmt := range strings.Split(strings.Join(kept, "\n"), ";") {
		if stmt = strings.TrimSpace(stmt); stmt != "" {
			out = append(out, stmt)
		}
	}
	return out
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
)

func openPostgres(c postgres.PostgresConfig) *gorm.DB {
	sslMode := "disable"
	if c.PostgresqlSSLMode {
		sslMode = "require"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.PostgresqlHost, c.PostgresqlUser, c.PostgresqlPassword, c.PostgresqlDbname, c.PostgresqlPort, sslMode)
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	return db
}
