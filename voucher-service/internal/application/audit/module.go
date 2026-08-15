package audit

import "go.uber.org/fx"

var Module = fx.Module("audit-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(AuditService)))),
)
