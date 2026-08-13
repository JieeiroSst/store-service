package webhook

import (
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"go.uber.org/fx"
)

var Module = fx.Module("webhook",
	fx.Provide(func() ports.WebhookSender { return NewHTTPSender() }),
)
