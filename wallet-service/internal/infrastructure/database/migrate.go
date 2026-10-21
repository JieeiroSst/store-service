package database

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const migrationLockID = 7_310_442_001

var baselines = []struct {
	version string
	table   string
}{
	{"001_init", "wallets"},
	{"002_wallet_features", "ledger_entries"},
}

func Migrate(ctx context.Context, db *sqlx.DB, schema fs.FS, log *zap.Logger) error {
	conn, err := db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrationLockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationLockID)
	}()

	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT        PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[string]bool{}
	var versions []string
	if err := conn.SelectContext(ctx, &versions, `SELECT version FROM schema_migrations`); err != nil {
		return err
	}
	for _, v := range versions {
		applied[v] = true
	}
	if len(applied) == 0 {
		for _, b := range baselines {
			var exists bool
			if err := conn.GetContext(ctx, &exists, `SELECT to_regclass('public.' || $1) IS NOT NULL`, b.table); err != nil {
				return err
			}
			if !exists {
				break
			}
			if _, err := conn.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, b.version); err != nil {
				return err
			}
			applied[b.version] = true
			log.Info("schema baselined from existing tables", zap.String("version", b.version))
		}
	}

	files, err := sqlFiles(schema, ".")
	if err != nil {
		return err
	}
	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			continue
		}
		body, err := fs.ReadFile(schema, name)
		if err != nil {
			return err
		}
		tx, err := conn.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Info("migration applied", zap.String("version", version))
	}
	return nil
}

func Seed(ctx context.Context, db *sqlx.DB, seed fs.FS, dir string, log *zap.Logger) error {
	files, err := sqlFiles(seed, dir)
	if err != nil {
		return err
	}
	for _, name := range files {
		body, err := fs.ReadFile(seed, name)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("seed %s: %w", name, err)
		}
		log.Info("seed applied", zap.String("file", name))
	}
	return nil
}

func sqlFiles(fsys fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			out = append(out, path.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}
