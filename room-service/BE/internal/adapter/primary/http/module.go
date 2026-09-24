package http

import (
	"github.com/JIeeiroSst/room-service/internal/adapter/primary/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(ws.NewHub, NewHandler, NewRouter),
)
