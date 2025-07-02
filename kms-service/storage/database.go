package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/kms/config"
	"github.com/JIeeiroSst/kms/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Database interface {
	SaveKey(*models.Key) error
	GetKey(string) (*models.Key, error)
	UpdateKey(*models.Key) error
	MarkKeyDeleted(string) error
	ListKeys() ([]models.Key, error)
	ListKeysByUser(uuid.UUID) ([]models.Key, error)
	IncrementKeyUseCount(string) error

	SaveAuditLog(models.AuditLog) error
	ListAuditLogs(int, int) ([]models.AuditLog, error)
	ListAuditLogsByUser(string, int, int) ([]models.AuditLog, error)
	GetAuditLogsByResource(string, int, int) ([]models.AuditLog, error)

	Close() error
}

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB() (*PostgresDB, error) {
	db, err := sql.Open("postgres", config.AppConfig.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	pgDB := &PostgresDB{db: db}

	if err := pgDB.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return pgDB, nil
}

func (p *PostgresDB) initSchema() error {
	schema := `
	-- Keys table
	CREATE TABLE IF NOT EXISTS keys (
		id UUID PRIMARY KEY,
		alias VARCHAR(255) NOT NULL UNIQUE,
		encrypted_key BYTEA NOT NULL,
		algorithm VARCHAR(50) NOT NULL,
		key_length INTEGER NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		last_rotated_at TIMESTAMP WITH TIME ZONE,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		version INTEGER NOT NULL DEFAULT 1,
		created_by UUID NOT NULL,
		tags JSONB,
		use_count BIGINT NOT NULL DEFAULT 0
	);
	
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		username VARCHAR(100) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL DEFAULT 'user',
		permissions TEXT[] NOT NULL DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT true
	);
	
	-- Audit logs table
	CREATE TABLE IF NOT EXISTS audit_logs (
		id UUID PRIMARY KEY,
		actor_id UUID,
		actor_name VARCHAR(255),
		action VARCHAR(255) NOT NULL,
		resource VARCHAR(100) NOT NULL,
		resource_id VARCHAR(255),
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
		ip_address INET,
		user_agent TEXT,
		success BOOLEAN NOT NULL,
		error_msg TEXT,
		metadata JSONB,
		level VARCHAR(20) NOT NULL DEFAULT 'info'
	);
	
	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_keys_created_by ON keys(created_by);
	CREATE INDEX IF NOT EXISTS idx_keys_status ON keys(status);
	CREATE INDEX IF NOT EXISTS idx_keys_expires_at ON keys(expires_at);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);
	`

	_, err := p.db.Exec(schema)
	return err
}

func (p *PostgresDB) SaveKey(k *models.Key) error {
	tagsJSON, _ := json.Marshal(k.Tags)

	_, err := p.db.Exec(`
		INSERT INTO keys (id, alias, encrypted_key, algorithm, key_length, created_at, updated_at, expires_at, status, version, created_by, tags, use_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		k.ID, k.Alias, k.EncryptedKey, k.Algorithm, k.KeyLength, k.CreatedAt, k.UpdatedAt, k.ExpiresAt, k.Status, k.Version, k.CreatedBy, tagsJSON, k.UseCount)
	return err
}

func (p *PostgresDB) GetKey(id string) (*models.Key, error) {
	row := p.db.QueryRow(`
		SELECT id, alias, encrypted_key, algorithm, key_length, created_at, updated_at, expires_at, last_rotated_at, status, version, created_by, tags, use_count
		FROM keys WHERE id=$1 AND status != 'deleted'`, id)

	k := &models.Key{}
	var tagsJSON []byte

	err := row.Scan(&k.ID, &k.Alias, &k.EncryptedKey, &k.Algorithm, &k.KeyLength, &k.CreatedAt, &k.UpdatedAt, &k.ExpiresAt, &k.LastRotatedAt, &k.Status, &k.Version, &k.CreatedBy, &tagsJSON, &k.UseCount)
	if err != nil {
		return nil, err
	}

	if tagsJSON != nil {
		json.Unmarshal(tagsJSON, &k.Tags)
	}

	return k, nil
}

func (p *PostgresDB) UpdateKey(k *models.Key) error {
	tagsJSON, _ := json.Marshal(k.Tags)

	_, err := p.db.Exec(`
		UPDATE keys SET encrypted_key=$1, updated_at=$2, expires_at=$3, last_rotated_at=$4, version=$5, tags=$6
		WHERE id=$7`,
		k.EncryptedKey, k.UpdatedAt, k.ExpiresAt, k.LastRotatedAt, k.Version, tagsJSON, k.ID)
	return err
}

func (p *PostgresDB) MarkKeyDeleted(id string) error {
	_, err := p.db.Exec(`UPDATE keys SET status='deleted', updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (p *PostgresDB) ListKeys() ([]models.Key, error) {
	rows, err := p.db.Query(`
		SELECT id, alias, algorithm, key_length, created_at, updated_at, expires_at, last_rotated_at, status, version, created_by, use_count
		FROM keys WHERE status != 'deleted' ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.Key
	for rows.Next() {
		var k models.Key
		err := rows.Scan(&k.ID, &k.Alias, &k.Algorithm, &k.KeyLength, &k.CreatedAt, &k.UpdatedAt, &k.ExpiresAt, &k.LastRotatedAt, &k.Status, &k.Version, &k.CreatedBy, &k.UseCount)
		if err != nil {
			continue
		}
		keys = append(keys, k)
	}

	return keys, nil
}

func (p *PostgresDB) ListKeysByUser(userID uuid.UUID) ([]models.Key, error) {
	rows, err := p.db.Query(`
		SELECT id, alias, algorithm, key_length, created_at, updated_at, expires_at, last_rotated_at, status, version, created_by, use_count
		FROM keys WHERE created_by=$1 AND status != 'deleted' ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.Key
	for rows.Next() {
		var k models.Key
		err := rows.Scan(&k.ID, &k.Alias, &k.Algorithm, &k.KeyLength, &k.CreatedAt, &k.UpdatedAt, &k.ExpiresAt, &k.LastRotatedAt, &k.Status, &k.Version, &k.CreatedBy, &k.UseCount)
		if err != nil {
			continue
		}
		keys = append(keys, k)
	}

	return keys, nil
}

func (p *PostgresDB) IncrementKeyUseCount(keyID string) error {
	_, err := p.db.Exec(`UPDATE keys SET use_count = use_count + 1 WHERE id = $1`, keyID)
	return err
}

func (p *PostgresDB) SaveAuditLog(log models.AuditLog) error {
	metadataJSON, _ := json.Marshal(log.Metadata)

	_, err := p.db.Exec(`
		INSERT INTO audit_logs (id, actor_id, actor_name, action, resource, resource_id, timestamp, ip_address, user_agent, success, error_msg, metadata, level)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		log.ID, log.ActorID, log.ActorName, log.Action, log.Resource, log.ResourceID, log.Timestamp, log.IPAddress, log.UserAgent, log.Success, log.ErrorMsg, metadataJSON, log.Level)
	return err
}

func (p *PostgresDB) ListAuditLogs(limit, offset int) ([]models.AuditLog, error) {
	rows, err := p.db.Query(`
		SELECT id, actor_id, actor_name, action, resource, resource_id, timestamp, ip_address, user_agent, success, error_msg, level
		FROM audit_logs ORDER BY timestamp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanAuditLogs(rows)
}

func (p *PostgresDB) ListAuditLogsByUser(userID string, limit, offset int) ([]models.AuditLog, error) {
	rows, err := p.db.Query(`
		SELECT id, actor_id, actor_name, action, resource, resource_id, timestamp, ip_address, user_agent, success, error_msg, level
		FROM audit_logs WHERE actor_id = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanAuditLogs(rows)
}

func (p *PostgresDB) GetAuditLogsByResource(resourceID string, limit, offset int) ([]models.AuditLog, error) {
	rows, err := p.db.Query(`
		SELECT id, actor_id, actor_name, action, resource, resource_id, timestamp, ip_address, user_agent, success, error_msg, level
		FROM audit_logs WHERE resource_id = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3`, resourceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanAuditLogs(rows)
}

func (p *PostgresDB) scanAuditLogs(rows *sql.Rows) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		err := rows.Scan(&l.ID, &l.ActorID, &l.ActorName, &l.Action, &l.Resource, &l.ResourceID, &l.Timestamp, &l.IPAddress, &l.UserAgent, &l.Success, &l.ErrorMsg, &l.Level)
		if err != nil {
			continue
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}
