package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPostRepository),
	fx.Provide(NewCommentRepository),
	fx.Provide(NewLikeRepository),
	fx.Provide(NewFollowRepository),
	fx.Provide(NewTagRepository),
	fx.Provide(NewBookmarkRepository),
)
