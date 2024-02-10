package logger

import (
	"encoding/json"
	"log"

	"go.uber.org/zap"
)

// sugar.Warn("this is a warning message")
func Logger() *zap.SugaredLogger {
	rawJSON := []byte(`{
        "level": "warn",
        "encoding": "json",
         "outputPaths": ["stdout"],
         "encoderConfig": {
            "levelKey": "level",
            "messageKey": "message",
            "levelEncoder": "lowercase"
        }
        }`)
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		log.Fatal(err)
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	return logger.Sugar()
}
