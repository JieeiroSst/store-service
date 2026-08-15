package consul

import "go.uber.org/fx"

var Module = fx.Module("consul", fx.Invoke(RegisterService))
