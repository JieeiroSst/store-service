package system

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(func() Clock { return Clock{} }, fx.As(new(port.Clock))),
		fx.Annotate(func() Random { return Random{} }, fx.As(new(port.Random))),
	),
)

type Clock struct{}

func (Clock) Now() time.Time { return time.Now().UTC() }

type Random struct{}

func (Random) Intn(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(v.Int64())
}
