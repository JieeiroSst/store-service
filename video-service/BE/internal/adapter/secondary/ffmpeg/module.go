package ffmpeg

import (
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(fx.Annotate(NewTranscoder, fx.As(new(port.Transcoder)))),
)
