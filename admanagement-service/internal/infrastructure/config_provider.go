package infrastructure

import "github.com/JIeeiroSst/admanagement-service/config"

func newConfig() *config.Config {
	return config.Load()
}
