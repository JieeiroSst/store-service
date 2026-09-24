package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func dsn(c config.PostgresConfig, db string) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, c.Port),
		Path:     "/" + db,
		RawQuery: url.Values{"sslmode": {c.SSLMode}}.Encode(),
	}
	return u.String()
}

var enums = []struct{ name, values string }{
	{"gender_type", "'Male', 'Female', 'Other'"},
	{"blood_type", "'A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-'"},
	{"appointment_status", "'Scheduled', 'Completed', 'Cancelled', 'No-Show'"},
}

func New(cfg *config.Config) (*gorm.DB, error) {
	c := cfg.Postgres

	if err := ensureDatabase(dsn(c, "postgres"), c.DBName); err != nil {
		return nil, fmt.Errorf("ensure database: %w", err)
	}

	db, err := gorm.Open(postgres.Open(dsn(c, c.DBName)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.Connection(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_lock(?)", migrationLockID).Error; err != nil {
			return err
		}
		defer tx.Exec("SELECT pg_advisory_unlock(?)", migrationLockID)
		return migrate(tx)
	})
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

const migrationLockID = 727_001

func migrate(db *gorm.DB) error {
	for _, e := range enums {
		stmt := fmt.Sprintf(`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = '%s') THEN
				CREATE TYPE %s AS ENUM (%s);
			END IF;
		END $$;`, e.name, e.name, e.values)
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	if err := db.AutoMigrate(model.Models()...); err != nil {
		return err
	}
	for _, fk := range foreignKeys {
		if err := addForeignKey(db, fk); err != nil {
			return err
		}
	}
	return nil
}

var foreignKeys = []struct{ table, column, refTable, refColumn string }{
	{"staff", "department_id", "departments", "department_id"},
	{"medical_records", "patient_id", "patients", "patient_id"},
	{"medical_records", "created_by", "staff", "staff_id"},
	{"appointments", "patient_id", "patients", "patient_id"},
	{"appointments", "staff_id", "staff", "staff_id"},
	{"appointments", "department_id", "departments", "department_id"},
	{"prescriptions", "patient_id", "patients", "patient_id"},
	{"prescriptions", "prescribed_by", "staff", "staff_id"},
	{"lab_results", "patient_id", "patients", "patient_id"},
	{"lab_results", "ordered_by", "staff", "staff_id"},
	{"billing", "patient_id", "patients", "patient_id"},
	{"billing", "appointment_id", "appointments", "appointment_id"},
	{"billing_accounts", "patient_id", "patients", "patient_id"},
}

func addForeignKey(db *gorm.DB, fk struct{ table, column, refTable, refColumn string }) error {
	name := fmt.Sprintf("fk_%s_%s", fk.table, fk.column)
	return db.Exec(fmt.Sprintf(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = '%[1]s') THEN
			ALTER TABLE %[2]s ADD CONSTRAINT %[1]s FOREIGN KEY (%[3]s) REFERENCES %[4]s (%[5]s);
		END IF;
	END $$;`, name, fk.table, fk.column, fk.refTable, fk.refColumn)).Error
}

func ensureDatabase(adminDSN, name string) error {
	conn, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var exists bool
	if err := conn.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = conn.ExecContext(ctx, `CREATE DATABASE "`+name+`"`)
	return err
}

func registerLifecycle(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
}

var Module = fx.Options(
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)
