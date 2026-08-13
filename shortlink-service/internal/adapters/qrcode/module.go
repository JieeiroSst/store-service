package qrcode

import (
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"go.uber.org/fx"
)

var Module = fx.Module("qrcode",
	fx.Provide(func() ports.QRCodeGenerator { return New() }),
)
