package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/auth"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// fileEnv gives the file tests a real (SQLite) file repository, so filtering,
// soft delete, versions and transactions behave as in production.
type fileEnv struct {
	db     *gorm.DB
	files  port.ContractFileRepository
	events port.Repository[model.ContractFileEvent]
	docs   docShim
}

func newFileEnv(tb testing.TB) *fileEnv {
	tb.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		tb.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(model.Models()...); err != nil {
		tb.Fatal(err)
	}
	e := &fileEnv{db: db, files: repository.NewFileRepository(db), events: repository.New[model.ContractFileEvent](db)}
	e.docs = docShim{db: db, repo: e.files}
	return e
}

// build makes the file service on top of the environment.
func (e *fileEnv) build(contracts port.Repository[model.Contract], store port.DocumentStore, tx port.TxRunner) *ContractFiles {
	return e.buildWith(contracts, store, tx, disabledScanner{}, config.FilesConfig{MaxBytes: 25 << 20})
}

func (e *fileEnv) buildWith(contracts port.Repository[model.Contract], store port.DocumentStore, tx port.TxRunner, sc port.Scanner, cfg config.FilesConfig) *ContractFiles {
	return NewContractFiles(ContractFilesDeps{
		Files: e.files, Events: e.events, Contracts: contracts, Store: store, Scanner: sc, Tx: tx,
		Config: &config.Config{Files: cfg},
	})
}

type disabledScanner struct{}

func (disabledScanner) Enabled() bool { return false }
func (disabledScanner) Scan(context.Context, string, []byte) (port.ScanResult, error) {
	return port.ScanResult{}, nil
}

// docShim is what the workflow tests use to look at stored files.
type docShim struct {
	db   *gorm.DB
	repo port.ContractFileRepository
}

func (d docShim) count() int {
	_, total, _ := d.repo.List(context.Background(), port.ListQuery{Limit: 1000})
	return int(total)
}

func (d docShim) row(id uint) *model.ContractFile {
	f, err := d.repo.GetByID(context.Background(), id)
	if err != nil {
		return nil
	}
	return f
}

func (d docShim) insert(f model.ContractFile) {
	if err := d.repo.Create(context.Background(), &f); err != nil {
		panic(err)
	}
}

func (d docShim) reset() { d.db.Exec("DELETE FROM crm_contract_file") }

var (
	viewer  = auth.Principal{Subject: "u-view", Name: "Vy", Roles: []auth.Role{auth.Viewer}}
	staff   = auth.Principal{Subject: "u-staff", Name: "Sam", Roles: []auth.Role{auth.Staff}, IP: "10.0.0.5"}
	staff2  = auth.Principal{Subject: "u-staff2", Name: "Sue", Roles: []auth.Role{auth.Staff}}
	manager = auth.Principal{Subject: "u-mgr", Name: "Max", Roles: []auth.Role{auth.Manager}}
	nobody  = auth.Principal{Subject: "u-none", Name: "Nia"}
)

func as(p auth.Principal) context.Context { return auth.WithPrincipal(context.Background(), p) }
