package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) MedicalCreateMedicine(ctx context.Context, input model.MedicalMedicineInput) (*model.MedicalMedicine, error) {
	return restCall[*model.MedicalMedicine](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: "/api/v1/medicines", Body: input}, "")
}

func (r *mutationResolver) MedicalReceiveBatch(ctx context.Context, id int, input model.MedicalBatchInput) (*model.MedicalBatch, error) {
	return restCall[*model.MedicalBatch](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/medicines/%v/batches", id), Body: withField(input, "received_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) MedicalAdjustBatch(ctx context.Context, id int, input model.MedicalAdjustInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/batches/%v/adjust", id), Body: withField(input, "adjusted_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) MedicalDisposeExpired(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: "/api/v1/inventory/dispose-expired", Body: withField(map[string]any{}, "disposed_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) MedicalCreateInteraction(ctx context.Context, input model.MedicalInteractionInput) (*model.MedicalInteractionRule, error) {
	return restCall[*model.MedicalInteractionRule](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: "/api/v1/interactions", Body: input}, "")
}

func (r *mutationResolver) MedicalAddAlias(ctx context.Context, alias string, canonical string) (*model.MedicalIngredientAlias, error) {
	return restCall[*model.MedicalIngredientAlias](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: "/api/v1/terminology/aliases", Body: map[string]any{"alias": alias, "canonical": canonical}}, "")
}

func (r *mutationResolver) MedicalAddAllergenGroup(ctx context.Context, ingredient string, allergenGroup string) (*model.MedicalIngredientAllergenGroup, error) {
	return restCall[*model.MedicalIngredientAllergenGroup](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: "/api/v1/terminology/allergen-groups", Body: map[string]any{"ingredient": ingredient, "allergen_group": allergenGroup}}, "")
}

func (r *mutationResolver) MedicalCreateDispense(ctx context.Context, input model.MedicalDispenseInput) (*model.MedicalDispense, error) {
	return restCall[*model.MedicalDispense](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: "/api/v1/dispenses", Body: withField(input, "dispensed_by", userIDOf(ctx))}, "")
}

func (r *queryResolver) MedicalMedicines(ctx context.Context, q *string, limit *int, offset *int) (*model.MedicalMedicineList, error) {
	return restCall[*model.MedicalMedicineList](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: "/api/v1/medicines", Query: restQuery("q", q, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) MedicalMedicine(ctx context.Context, id int) (*model.MedicalMedicine, error) {
	return restCall[*model.MedicalMedicine](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/medicines/%v", id)}, "")
}

func (r *queryResolver) MedicalStock(ctx context.Context, id int) (*model.MedicalStockSummary, error) {
	return restCall[*model.MedicalStockSummary](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/medicines/%v/stock", id)}, "")
}

func (r *queryResolver) MedicalMovements(ctx context.Context, id int, limit *int, offset *int) ([]*model.MedicalStockMovement, error) {
	return restCall[[]*model.MedicalStockMovement](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/medicines/%v/movements", id), Query: restQuery("limit", limit, "offset", offset)}, "items")
}

func (r *queryResolver) MedicalExpiring(ctx context.Context, days *int) ([]*model.MedicalBatch, error) {
	return restCall[[]*model.MedicalBatch](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: "/api/v1/inventory/expiring", Query: restQuery("days", days)}, "items")
}

func (r *queryResolver) MedicalLowStock(ctx context.Context) ([]*model.MedicalLowStockItem, error) {
	return restCall[[]*model.MedicalLowStockItem](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: "/api/v1/inventory/low-stock"}, "items")
}

func (r *queryResolver) MedicalInteractions(ctx context.Context, limit *int, offset *int) ([]*model.MedicalInteractionRule, error) {
	return restCall[[]*model.MedicalInteractionRule](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: "/api/v1/interactions", Query: restQuery("limit", limit, "offset", offset)}, "items")
}

func (r *queryResolver) MedicalTerminology(ctx context.Context) (*model.MedicalTerminologyList, error) {
	return restCall[*model.MedicalTerminologyList](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: "/api/v1/terminology"}, "")
}

func (r *queryResolver) MedicalDispenses(ctx context.Context, patientRef *string, limit *int, offset *int) ([]*model.MedicalDispense, error) {
	return restCall[[]*model.MedicalDispense](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: "/api/v1/dispenses", Query: restQuery("patient_ref", patientRef, "limit", limit, "offset", offset)}, "items")
}

func (r *queryResolver) MedicalDispense(ctx context.Context, id int) (*model.MedicalDispense, error) {
	return restCall[*model.MedicalDispense](ctx, r.Proxy, "medical-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/dispenses/%v", id)}, "")
}

func (r *queryResolver) MedicalCheckSafety(ctx context.Context, input model.MedicalSafetyCheckInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "medical-service", rest.Request{Method: "POST", Path: "/api/v1/interactions/check", Body: input}, "")
}
