package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) MediaQuery() generated.MediaQueryResolver { return &mediaQueryResolver{r} }

type mediaQueryResolver struct{ *Resolver }

func (r *mediaQueryResolver) Video(ctx context.Context, obj *model.MediaQuery, videoID int, userID int) (*model.MediaVideo, error) {
	return r.Clients.MediaService.Video(ctx, videoID, userID)
}
