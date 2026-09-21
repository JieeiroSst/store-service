package restaurantclient

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewRestaurantClient),
)
