package application

import (
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/config"
)

var (
	NotificationRetryDelays = []time.Duration{5 * time.Second, 30 * time.Second, 2 * time.Minute, 10 * time.Minute}
	BatchRetryDelays        = []time.Duration{10 * time.Second, time.Minute, 5 * time.Minute, 15 * time.Minute}
)

type CampaignSettings struct {
	PushBatchSize    int
	EmailBatchSize   int
	MaxRecipients    int
	MaxAttempts      int
	RatePerSecond    float64
	PlannerLease     time.Duration
	BatchLease       time.Duration
	RepublishGrace   time.Duration
	BatchesPerPlan   int
	RepublishPerTick int
}

func NewCampaignSettings(cfg *config.Config) CampaignSettings {
	s := CampaignSettings{
		PushBatchSize:    500,
		EmailBatchSize:   100,
		MaxRecipients:    1_000_000,
		MaxAttempts:      len(BatchRetryDelays) + 1,
		PlannerLease:     2 * time.Minute,
		BatchLease:       5 * time.Minute,
		RepublishGrace:   10 * time.Minute,
		BatchesPerPlan:   100,
		RepublishPerTick: 500,
	}
	c := cfg.Campaign
	if c.PushBatchSize > 0 && c.PushBatchSize < s.PushBatchSize {
		s.PushBatchSize = c.PushBatchSize
	}
	if c.EmailBatchSize > 0 && c.EmailBatchSize < s.EmailBatchSize {
		s.EmailBatchSize = c.EmailBatchSize
	}
	if c.MaxRecipients > 0 {
		s.MaxRecipients = c.MaxRecipients
	}
	if c.MaxAttempts > 0 && c.MaxAttempts <= len(BatchRetryDelays)+1 {
		s.MaxAttempts = c.MaxAttempts
	}
	if c.RatePerSecond > 0 {
		s.RatePerSecond = c.RatePerSecond
	}
	return s
}

func retryDelay(delays []time.Duration, attempt int) time.Duration {
	if attempt < 1 {
		return delays[0]
	}
	return delays[min(attempt, len(delays))-1]
}
