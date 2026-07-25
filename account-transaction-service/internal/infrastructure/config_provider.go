package infrastructure

import (
	"github.com/Jieeirosst/account-transaction-service/config"
)

func newConfig() (*config.Config, error) {
	return config.InitializeConfiguration(".env")
}
