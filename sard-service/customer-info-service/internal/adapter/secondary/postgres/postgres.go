package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

const (
	migrationLockID = 7302_0001
	uniqueViolation = "23505"
)

type txKey struct{}

type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type DB struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string, maxConns int32) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() { db.pool.Close() }

func (db *DB) Ping(ctx context.Context) error { return db.pool.Ping(ctx) }

func (db *DB) Migrate(ctx context.Context) error {
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", int64(migrationLockID)); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, schema)
		return err
	})
}

func (db *DB) q(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db.pool
}

func (db *DB) WithLock(ctx context.Context, key string, fn func(ctx context.Context) error) error {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockID(key)); err != nil {
			return err
		}
		return fn(ctx)
	}
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockID(key)); err != nil {
			return err
		}
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

func lockID(key string) int64 {
	h := fnv.New64a()
	h.Write([]byte("customer-info/" + key))
	return int64(h.Sum64())
}

type Customers struct{ db *DB }

func NewCustomers(db *DB) Customers { return Customers{db: db} }

const columns = `id, user_id, username, full_name, email, phone, address, gender, status,
	kyc_status, kyc_document_hash, kyc, created_at, updated_at, synced_at`

type kycRecord struct {
	DocumentType  string     `json:"document_type,omitempty"`
	Source        string     `json:"source,omitempty"`
	Number        string     `json:"number,omitempty"`
	FullName      string     `json:"full_name,omitempty"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	Gender        string     `json:"gender,omitempty"`
	Nationality   string     `json:"nationality,omitempty"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`
	ChecksumValid bool       `json:"checksum_valid,omitempty"`
	NFCVerified   bool       `json:"nfc_verified,omitempty"`
	FaceVerified  bool       `json:"face_verified,omitempty"`
	Confidence    float64    `json:"confidence,omitempty"`
	Reasons       []string   `json:"reasons,omitempty"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
}

func toRecord(k domain.KYC) kycRecord {
	d := k.Document
	return kycRecord{
		DocumentType: k.DocumentType, Source: d.Source, Number: d.Last4(), FullName: k.FullName,
		DateOfBirth: d.DateOfBirth, Gender: d.Gender, Nationality: d.Nationality, ExpiryDate: d.ExpiryDate,
		ChecksumValid: d.ChecksumValid, NFCVerified: d.NFCVerified, FaceVerified: k.FaceVerified,
		Confidence: d.Confidence, Reasons: k.Reasons, VerifiedAt: k.VerifiedAt, SubmittedAt: k.SubmittedAt,
	}
}

func fromRecord(status, hash string, r kycRecord) domain.KYC {
	return domain.KYC{
		Status:       domain.KYCStatus(status),
		DocumentType: r.DocumentType,
		DocumentHash: hash,
		FullName:     r.FullName,
		FaceVerified: r.FaceVerified,
		Reasons:      r.Reasons,
		VerifiedAt:   r.VerifiedAt,
		SubmittedAt:  r.SubmittedAt,
		Document: domain.IdentityDocument{
			Source: r.Source, Number: r.Number, FullName: r.FullName, DateOfBirth: r.DateOfBirth, Gender: r.Gender,
			Nationality: r.Nationality, ExpiryDate: r.ExpiryDate, ChecksumValid: r.ChecksumValid,
			NFCVerified: r.NFCVerified, Confidence: r.Confidence,
		},
	}
}

func (r Customers) Create(ctx context.Context, c *domain.Customer) error {
	kyc, err := json.Marshal(toRecord(c.KYC))
	if err != nil {
		return err
	}
	_, err = r.db.q(ctx).Exec(ctx, `INSERT INTO customers (`+columns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		c.ID, c.UserID, c.Username, c.FullName, c.Email, c.Phone, c.Address, c.Gender, string(c.Status),
		string(c.KYC.Status), c.KYC.DocumentHash, kyc, c.CreatedAt, c.UpdatedAt, c.SyncedAt)
	return mapErr(err)
}

func (r Customers) Update(ctx context.Context, c *domain.Customer) error {
	kyc, err := json.Marshal(toRecord(c.KYC))
	if err != nil {
		return err
	}
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE customers SET
		username = $2, full_name = $3, email = $4, phone = $5, address = $6, gender = $7, status = $8,
		kyc_status = $9, kyc_document_hash = $10, kyc = $11, updated_at = $12, synced_at = $13
		WHERE id = $1`,
		c.ID, c.Username, c.FullName, c.Email, c.Phone, c.Address, c.Gender, string(c.Status),
		string(c.KYC.Status), c.KYC.DocumentHash, kyc, c.UpdatedAt, c.SyncedAt)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r Customers) Get(ctx context.Context, id string) (*domain.Customer, error) {
	return scan(r.db.q(ctx).QueryRow(ctx, `SELECT `+columns+` FROM customers WHERE id = $1`, id))
}

func (r Customers) GetByUserID(ctx context.Context, userID int64) (*domain.Customer, error) {
	return scan(r.db.q(ctx).QueryRow(ctx, `SELECT `+columns+` FROM customers WHERE user_id = $1`, userID))
}

func (r Customers) GetByDocumentHash(ctx context.Context, hash string) (*domain.Customer, error) {
	return scan(r.db.q(ctx).QueryRow(ctx, `SELECT `+columns+` FROM customers
		WHERE kyc_status = 'VERIFIED' AND kyc_document_hash = $1`, hash))
}

func scan(row pgx.Row) (*domain.Customer, error) {
	var (
		c                 domain.Customer
		status, kycStatus string
		hash              string
		raw               []byte
	)
	err := row.Scan(&c.ID, &c.UserID, &c.Username, &c.FullName, &c.Email, &c.Phone, &c.Address, &c.Gender,
		&status, &kycStatus, &hash, &raw, &c.CreatedAt, &c.UpdatedAt, &c.SyncedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var rec kycRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, fmt.Errorf("decode kyc of customer %s: %w", c.ID, err)
	}
	c.Status = domain.CustomerStatus(status)
	c.KYC = fromRecord(kycStatus, hash, rec)
	c.CreatedAt, c.UpdatedAt, c.SyncedAt = c.CreatedAt.UTC(), c.UpdatedAt.UTC(), c.SyncedAt.UTC()
	return &c, nil
}

func mapErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		if pgErr.ConstraintName == "customers_verified_document_idx" {
			return domain.Conflict("this identity document is already verified for another customer")
		}
		return domain.Conflict("customer already exists")
	}
	return err
}
