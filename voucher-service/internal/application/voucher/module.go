package voucher

import "go.uber.org/fx"

var Module = fx.Module("voucher-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(VoucherService)))),
)
