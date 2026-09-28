package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
)

type MedicineRepository interface {
	Create(ctx context.Context, m *model.Medicine) error
	GetByID(ctx context.Context, id int64) (*model.Medicine, error)

	GetByIDs(ctx context.Context, ids []int64) ([]model.Medicine, error)
	List(ctx context.Context, query string, limit, offset int) ([]model.Medicine, int64, error)
	ListLowStock(ctx context.Context, today model.Date) ([]model.LowStockItem, error)
}

type BatchRepository interface {
	Create(ctx context.Context, b *model.Batch) error
	Update(ctx context.Context, b *model.Batch) error
	GetByIDForUpdate(ctx context.Context, id int64) (*model.Batch, error)
	FindByNumberForUpdate(ctx context.Context, medicineID int64, batchNumber string) (*model.Batch, error)
	ListByMedicine(ctx context.Context, medicineID int64) ([]model.Batch, error)
	ListUsableForUpdate(ctx context.Context, medicineID int64, today model.Date) ([]model.Batch, error)
	ListExpiredForUpdate(ctx context.Context, today model.Date) ([]model.Batch, error)
	ListExpiring(ctx context.Context, today, until model.Date) ([]model.Batch, error)
}

type MovementRepository interface {
	Create(ctx context.Context, movements []model.StockMovement) error
	ListByMedicine(ctx context.Context, medicineID int64, limit, offset int) ([]model.StockMovement, error)
}

type InteractionRepository interface {
	Upsert(ctx context.Context, r *model.InteractionRule) error
	List(ctx context.Context, limit, offset int) ([]model.InteractionRule, error)

	FindAmong(ctx context.Context, ingredients []string) ([]model.InteractionRule, error)
}

type DispenseRepository interface {
	Create(ctx context.Context, d *model.Dispense) error
	GetByID(ctx context.Context, id int64) (*model.Dispense, error)
	ListByPatient(ctx context.Context, patientRef string, limit, offset int) ([]model.Dispense, error)
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time

	Today() model.Date
}

type TerminologyRepository interface {
	Load(ctx context.Context) (model.Terminology, error)

	UpsertAlias(ctx context.Context, a *model.IngredientAlias) error

	AddAllergenGroup(ctx context.Context, g *model.IngredientAllergenGroup) error
	ListAliases(ctx context.Context) ([]model.IngredientAlias, error)
	ListAllergenGroups(ctx context.Context) ([]model.IngredientAllergenGroup, error)
}
