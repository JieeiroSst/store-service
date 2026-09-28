package infrastructure

import (
	"fmt"
	"time"

	"github.com/JIeeiroSst/medical-service/config"
	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

type clock struct {
	loc *time.Location
}

func newClock(cfg *config.Config) (port.Clock, error) {
	loc, err := time.LoadLocation(cfg.Pharmacy.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("pharmacy time zone %q: %w", cfg.Pharmacy.TimeZone, err)
	}
	return &clock{loc: loc}, nil
}

func (c *clock) Now() time.Time    { return time.Now() }
func (c *clock) Today() model.Date { return model.DateOf(time.Now().In(c.loc)) }
