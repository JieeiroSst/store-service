package cron

import (
	"github.com/JIeeiroSst/coupon-service/internal/usecase"
	"github.com/robfig/cron"
)

type Cron struct {
	usecase usecase.Usecase
}

func NewCron(usecase usecase.Usecase) *Cron {
	return &Cron{usecase: usecase}
}

func (c *Cron) Run() *cron.Cron {
	cron := cron.New()
	cron.AddFunc("@every 1m", func() {})
	cron.Start()
	return cron

}
