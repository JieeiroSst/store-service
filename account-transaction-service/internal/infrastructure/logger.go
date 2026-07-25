package infrastructure

import "github.com/JIeeiroSst/utils/logger"

func initLogger() {
	logger.InitDefault(logger.Config{
		Level:      "info",
		JSONFormat: true,
		AppName:    "account-transaction-service",
	})
}
