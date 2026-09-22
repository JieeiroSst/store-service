package scheduler

import (
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/JIeeiroSst/partner-service/internal/logger"
	"github.com/robfig/cron"
)

type PartnerScheduler struct {
	cron *cron.Cron
	svc  services.PartnerService
}

func NewPartnerScheduler(svc services.PartnerService) *PartnerScheduler {
	return &PartnerScheduler{
		cron: cron.New(),
		svc:  svc,
	}
}

func (s *PartnerScheduler) Start() error {
	if err := s.cron.AddFunc("@daily", s.closeInactivePartners); err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

func (s *PartnerScheduler) Stop() {
	s.cron.Stop()
}

func (s *PartnerScheduler) closeInactivePartners() {
	closed, err := s.svc.CloseInactivePartners()
	if err != nil {
		logger.Log.Errorf("close inactive partners: %v", err)
		return
	}
	if closed > 0 {
		logger.Log.Infof("closed %d inactive partner(s)", closed)
	}
}
