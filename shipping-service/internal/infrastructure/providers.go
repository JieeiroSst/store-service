package infrastructure

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/application"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

func newClock() port.Clock { return clock{} }

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type codeGenerator struct{}

func (codeGenerator) NewShipmentCode() string {
	b := make([]byte, 8)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			panic(err)
		}
		b[i] = codeAlphabet[n.Int64()]
	}
	return "SHP" + time.Now().UTC().Format("060102") + string(b)
}

func newCodeGenerator() port.CodeGenerator { return codeGenerator{} }

func newSettings(cfg *config.Config) application.Settings {
	s := application.DefaultSettings()
	s.MaxCODAmount = cfg.Shipping.MaxCOD()
	return s
}
