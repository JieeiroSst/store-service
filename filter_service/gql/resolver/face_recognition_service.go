package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) FaceRecQuery() generated.FaceRecQueryResolver { return &faceRecQueryResolver{r} }

type faceRecQueryResolver struct{ *Resolver }

func (r *faceRecQueryResolver) FacePosition(ctx context.Context, obj *model.FaceRecQuery) (*model.FaceRecPositions, error) {
	return r.Clients.FaceRecognitionService.FacePosition(ctx)
}
