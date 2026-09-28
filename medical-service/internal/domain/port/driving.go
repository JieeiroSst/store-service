package port

import (
	"context"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
)

type ListInput struct {
	Limit  int
	Offset int
}

type CreateMedicineInput struct {
	Code                 string
	Name                 string
	Ingredients          []string
	AllergenGroups       []string
	DosageForm           string
	Strength             string
	Unit                 string
	RequiresPrescription bool
	ReorderLevel         int
}

type MedicineList struct {
	Items []model.Medicine `json:"items"`
	Total int64            `json:"total"`
}

type MedicineUsecase interface {
	CreateMedicine(ctx context.Context, in CreateMedicineInput) (*model.Medicine, error)
	GetMedicine(ctx context.Context, id int64) (*model.Medicine, error)
	ListMedicines(ctx context.Context, query string, in ListInput) (*MedicineList, error)
}

type ReceiveBatchInput struct {
	MedicineID  int64
	BatchNumber string
	ExpiryDate  model.Date
	Quantity    int
	UnitCost    int64
	Supplier    string
	ReceivedBy  string
}

type AdjustBatchInput struct {
	BatchID     int64
	NewQuantity int
	Reason      string
	AdjustedBy  string
}

type DisposeResult struct {
	Batches int `json:"batches"`
	Units   int `json:"units"`
}

type InventoryUsecase interface {
	ReceiveBatch(ctx context.Context, in ReceiveBatchInput) (*model.Batch, error)
	AdjustBatch(ctx context.Context, in AdjustBatchInput) (*model.Batch, error)
	GetStock(ctx context.Context, medicineID int64) (*model.StockSummary, error)
	ListExpiring(ctx context.Context, withinDays int) ([]model.Batch, error)
	ListLowStock(ctx context.Context) ([]model.LowStockItem, error)
	DisposeExpired(ctx context.Context, disposedBy string) (*DisposeResult, error)
	ListMovements(ctx context.Context, medicineID int64, in ListInput) ([]model.StockMovement, error)
}

type CreateInteractionInput struct {
	IngredientA string
	IngredientB string
	Severity    model.Severity
	Description string
}

type SafetyCheckInput struct {
	MedicineIDs        []int64
	CurrentMedicineIDs []int64
	Allergies          []string
}

type InteractionUsecase interface {
	CreateRule(ctx context.Context, in CreateInteractionInput) (*model.InteractionRule, error)
	ListRules(ctx context.Context, in ListInput) ([]model.InteractionRule, error)
	Check(ctx context.Context, in SafetyCheckInput) (*model.SafetyReport, error)
}

type DispenseItemInput struct {
	MedicineID int64
	Quantity   int
}

type DispenseInput struct {
	PatientRef         string
	PrescriptionRef    string
	DispensedBy        string
	Items              []DispenseItemInput
	CurrentMedicineIDs []int64
	Allergies          []string
	OverrideReason     string
}

type DispenseUsecase interface {
	Dispense(ctx context.Context, in DispenseInput) (*model.Dispense, error)
	GetDispense(ctx context.Context, id int64) (*model.Dispense, error)
	ListByPatient(ctx context.Context, patientRef string, in ListInput) ([]model.Dispense, error)
}

type TerminologyList struct {
	Aliases        []model.IngredientAlias         `json:"aliases"`
	AllergenGroups []model.IngredientAllergenGroup `json:"allergen_groups"`
}

type TerminologyUsecase interface {
	AddAlias(ctx context.Context, alias, canonical string) (*model.IngredientAlias, error)
	AddAllergenGroup(ctx context.Context, ingredient, group string) (*model.IngredientAllergenGroup, error)
	List(ctx context.Context) (*TerminologyList, error)
}
