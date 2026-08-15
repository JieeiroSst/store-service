package authtoken

import (
	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	"go.uber.org/fx"
)

var Module = fx.Module("authtoken",
	fx.Provide(fx.Annotate(NewJWTIssuer, fx.As(new(authapp.TokenIssuer)))),
)
