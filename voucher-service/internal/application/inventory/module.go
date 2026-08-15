package inventory

import "go.uber.org/fx"

var Module = fx.Module("inventory-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(InventoryService)))),
)
