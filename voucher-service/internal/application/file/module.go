package file

import "go.uber.org/fx"

var Module = fx.Module("file-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(FileService)))),
)
