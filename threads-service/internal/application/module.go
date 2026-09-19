package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPostService),
	fx.Provide(NewCommentService),
	fx.Provide(NewLikeService),
	fx.Provide(NewFollowService),
	fx.Provide(NewBookmarkService),
	fx.Provide(NewTagService),
)
