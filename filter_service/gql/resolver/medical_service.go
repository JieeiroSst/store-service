package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) MedicalQuery() generated.MedicalQueryResolver { return &medicalQueryResolver{r} }

type medicalQueryResolver struct{ *Resolver }

func (r *medicalQueryResolver) Medicines(ctx context.Context, obj *model.MedicalQuery, qArg *string, limit *int, offset *int) (*model.MedicalMedicineList, error) {
	return r.Clients.MedicalService.Medicines(ctx, qArg, limit, offset)
}

func (r *medicalQueryResolver) Medicine(ctx context.Context, obj *model.MedicalQuery, id int) (*model.MedicalMedicine, error) {
	return r.Clients.MedicalService.Medicine(ctx, id)
}

func (r *medicalQueryResolver) Stock(ctx context.Context, obj *model.MedicalQuery, id int) (*model.MedicalStockSummary, error) {
	return r.Clients.MedicalService.Stock(ctx, id)
}

func (r *medicalQueryResolver) Movements(ctx context.Context, obj *model.MedicalQuery, id int, limit *int, offset *int) ([]*model.MedicalStockMovement, error) {
	return r.Clients.MedicalService.Movements(ctx, id, limit, offset)
}

func (r *medicalQueryResolver) Expiring(ctx context.Context, obj *model.MedicalQuery, days *int) ([]*model.MedicalBatch, error) {
	return r.Clients.MedicalService.Expiring(ctx, days)
}

func (r *medicalQueryResolver) LowStock(ctx context.Context, obj *model.MedicalQuery) ([]*model.MedicalLowStockItem, error) {
	return r.Clients.MedicalService.LowStock(ctx)
}

func (r *medicalQueryResolver) Interactions(ctx context.Context, obj *model.MedicalQuery, limit *int, offset *int) ([]*model.MedicalInteractionRule, error) {
	return r.Clients.MedicalService.Interactions(ctx, limit, offset)
}

func (r *medicalQueryResolver) Terminology(ctx context.Context, obj *model.MedicalQuery) (*model.MedicalTerminologyList, error) {
	return r.Clients.MedicalService.Terminology(ctx)
}

func (r *medicalQueryResolver) Dispenses(ctx context.Context, obj *model.MedicalQuery, patientRef *string, limit *int, offset *int) ([]*model.MedicalDispense, error) {
	return r.Clients.MedicalService.Dispenses(ctx, patientRef, limit, offset)
}

func (r *medicalQueryResolver) Dispense(ctx context.Context, obj *model.MedicalQuery, id int) (*model.MedicalDispense, error) {
	return r.Clients.MedicalService.Dispense(ctx, id)
}
