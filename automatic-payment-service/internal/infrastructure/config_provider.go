package infrastructure

import (
	"github.com/JIeeiroSst/automatic-payment-service/config"
)

func newConfig() (*config.Config, error) {
	return config.InitializeConfiguration(".env")
}
