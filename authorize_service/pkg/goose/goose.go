package goose

import (
	"database/sql"
	"embed"

	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/pressly/goose/v3"
)

type Migration struct {
	db *sql.DB
}

var embedMigrations embed.FS

func NewMigration(db *sql.DB) *Migration {
	return &Migration{
		db: db,
	}
}

func (m *Migration) RunMigration() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		log.Error(err.Error())
		return err
	}

	err := goose.Up(m.db, "migrations", goose.WithAllowMissing())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
