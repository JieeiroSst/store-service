package infrastructure

import (
	"net/http"
	"time"

	"github.com/JIeeiroSst/calculate-service/config"
)

func NewHTTPClient(cfg *config.Config) *http.Client {
	timeout := time.Duration(cfg.Weather.RequestTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &http.Client{Timeout: timeout}
}
