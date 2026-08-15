//go:build integration

package postgres_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/adapters/outbound/postgres"
	"github.com/JIeeiroSst/voucher-service/internal/domain/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("voucher_test"),
		tcpostgres.WithUsername("voucher"),
		tcpostgres.WithPassword("voucher"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	applyMigrations(t, db)
	return db
}

func TestVoucherRepository_OptimisticLocking_ConcurrentSaveIsRejected(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	merchantRepo := postgres.NewMerchantRepository(db)
	voucherRepo := postgres.NewVoucherRepository(db)

	now := time.Now().UTC()
	m, err := merchant.NewMerchant("Integration Test Merchant", shared.ProviderTypeSelf, nil, now)
	require.NoError(t, err)
	require.NoError(t, merchantRepo.Create(ctx, m))

	ref := shared.ProductRef{MerchantID: m.ID, SKU: "SKU-IT", Denomination: shared.NewMoney(10000, "VND")}
	v, err := voucher.NewVoucher(m.ID, ref, now)
	require.NoError(t, err)
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "ITCODE1"}, nil, now))
	require.NoError(t, voucherRepo.Create(ctx, v))

	// Two independent reads of the same row, simulating two concurrent
	// requests each about to mutate and save.
	first, err := voucherRepo.FindByID(ctx, v.ID)
	require.NoError(t, err)
	second, err := voucherRepo.FindByID(ctx, v.ID)
	require.NoError(t, err)
	require.Equal(t, first.PersistedVersion, second.PersistedVersion)

	require.NoError(t, first.Activate(voucher.OwnerTypeUser, "user-1", now))
	require.NoError(t, voucherRepo.Save(ctx, first))

	// second still thinks the row is at the version it read before
	// `first` committed its change — its Save must be rejected rather
	// than silently overwriting first's write.
	require.NoError(t, second.Activate(voucher.OwnerTypeUser, "user-2", now))
	err = voucherRepo.Save(ctx, second)
	require.ErrorIs(t, err, voucher.ErrVersionConflict)

	// The persisted row must reflect only `first`'s write.
	reloaded, err := voucherRepo.FindByID(ctx, v.ID)
	require.NoError(t, err)
	require.Equal(t, voucher.OwnerTypeUser, reloaded.OwnerType)
	require.Equal(t, "user-1", *reloaded.OwnerID)
}

func TestVoucherRepository_FindByIDForUpdate_LocksRowForConcurrentReader(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	txMgr := txmanager.NewGormTxManager(db)

	merchantRepo := postgres.NewMerchantRepository(db)
	voucherRepo := postgres.NewVoucherRepository(db)

	now := time.Now().UTC()
	m, err := merchant.NewMerchant("Integration Test Merchant 2", shared.ProviderTypeSelf, nil, now)
	require.NoError(t, err)
	require.NoError(t, merchantRepo.Create(ctx, m))

	ref := shared.ProductRef{MerchantID: m.ID, SKU: "SKU-IT2", Denomination: shared.NewMoney(10000, "VND")}
	v, err := voucher.NewVoucher(m.ID, ref, now)
	require.NoError(t, err)
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "ITCODE2"}, nil, now))
	require.NoError(t, voucherRepo.Create(ctx, v))

	lockAcquired := make(chan struct{})
	secondReaderProceeded := make(chan time.Time, 1)
	firstTxDone := make(chan time.Time, 1)

	go func() {
		_ = txMgr.WithinTx(ctx, func(txCtx context.Context) error {
			if _, err := voucherRepo.FindByIDForUpdate(txCtx, v.ID); err != nil {
				return err
			}
			close(lockAcquired)
			time.Sleep(300 * time.Millisecond)
			firstTxDone <- time.Now()
			return nil
		})
	}()

	<-lockAcquired

	go func() {
		_ = txMgr.WithinTx(ctx, func(txCtx context.Context) error {
			_, err := voucherRepo.FindByIDForUpdate(txCtx, v.ID)
			secondReaderProceeded <- time.Now()
			return err
		})
	}()

	doneAt := <-firstTxDone
	proceededAt := <-secondReaderProceeded
	require.False(t, proceededAt.Before(doneAt), "second FOR UPDATE reader must not proceed before the first transaction released its lock")
}

func applyMigrations(t *testing.T, db *gorm.DB) {
	t.Helper()
	root, err := findRepoRoot()
	require.NoError(t, err)
	sqlBytes, err := os.ReadFile(filepath.Join(root, "migrations", "000001_init.up.sql"))
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(sqlBytes)).Error)
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
