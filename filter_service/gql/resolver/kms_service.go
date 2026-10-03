package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) KmsQuery() generated.KmsQueryResolver { return &kmsQueryResolver{r} }

type kmsQueryResolver struct{ *Resolver }

func (r *kmsQueryResolver) Keys(ctx context.Context, obj *model.KmsQuery) ([]*model.KmsKey, error) {
	return r.Clients.KmsService.Keys(ctx)
}

func (r *kmsQueryResolver) Key(ctx context.Context, obj *model.KmsQuery, id string) (*model.KmsKey, error) {
	return r.Clients.KmsService.Key(ctx, id)
}

func (r *kmsQueryResolver) KeyForUse(ctx context.Context, obj *model.KmsQuery, id string) (*model.KmsKeyUse, error) {
	return r.Clients.KmsService.KeyForUse(ctx, id)
}

func (r *kmsQueryResolver) KeyUsageStats(ctx context.Context, obj *model.KmsQuery, id string) (*model.KmsKeyStats, error) {
	return r.Clients.KmsService.KeyUsageStats(ctx, id)
}

func (r *kmsQueryResolver) KeyAuditLogs(ctx context.Context, obj *model.KmsQuery, id string, limit *int, offset *int) (*model.KmsKeyAuditLogs, error) {
	return r.Clients.KmsService.KeyAuditLogs(ctx, id, limit, offset)
}

func (r *kmsQueryResolver) AuditLogs(ctx context.Context, obj *model.KmsQuery, limit *int, offset *int) (*model.KmsAuditLogs, error) {
	return r.Clients.KmsService.AuditLogs(ctx, limit, offset)
}
