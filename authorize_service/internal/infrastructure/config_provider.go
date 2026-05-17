package infrastructure

import (
	"github.com/JieeiroSst/authorize-service/config"
)

func newConfig() (*config.Config, error) {
	return config.InitializeConfiguration(".env")
}
