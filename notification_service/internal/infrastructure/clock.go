package infrastructure

import (
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

func newClock() port.Clock { return clock{} }
