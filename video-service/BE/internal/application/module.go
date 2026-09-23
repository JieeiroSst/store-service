package application

import "go.uber.org/fx"

// APIModule wires the use cases the HTTP API needs.
var APIModule = fx.Options(
	fx.Provide(NewVideoService),
)

// WorkerModule wires the use case the transcode worker needs.
var WorkerModule = fx.Options(
	fx.Provide(NewTranscodeService),
)
