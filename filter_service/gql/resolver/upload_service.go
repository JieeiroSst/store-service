package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) UploadQuery() generated.UploadQueryResolver { return &uploadQueryResolver{r} }

type uploadQueryResolver struct{ *Resolver }

func (r *uploadQueryResolver) Files(ctx context.Context, obj *model.UploadQuery, receiverID *string, limit *int, offset *int) (*model.UploadFileList, error) {
	return r.Clients.UploadService.Files(ctx, receiverID, limit, offset)
}

func (r *uploadQueryResolver) File(ctx context.Context, obj *model.UploadQuery, id string) (*model.UploadFile, error) {
	return r.Clients.UploadService.File(ctx, id)
}
