package provider

import (
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"go.uber.org/fx"
)

var Module = fx.Module("provider",
	fx.Provide(
		fx.Annotate(NewAPIProvider,
			fx.As(new(voucherapp.MerchantProvider)),
			fx.ResultTags(`group:"providers"`)),
		fx.Annotate(NewStockProvider,
			fx.As(new(voucherapp.MerchantProvider)),
			fx.ResultTags(`group:"providers"`)),
		fx.Annotate(NewSelfProvider,
			fx.As(new(voucherapp.MerchantProvider)),
			fx.ResultTags(`group:"providers"`)),
		fx.Annotate(NewRegistry,
			fx.ParamTags(`group:"providers"`),
			fx.As(new(voucherapp.ProviderRegistry))),
	),
)
