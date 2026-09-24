package usecase

import (
	"math"
	"time"
)

type PricingPolicy struct {
	LateGrace      time.Duration
	LateMultiplier float64
}

func (p PricingPolicy) withDefaults() PricingPolicy {
	if p.LateMultiplier <= 0 {
		p.LateMultiplier = 1
	}
	return p
}

func (p PricingPolicy) lateFee(overdue time.Duration, hourly float64) float64 {
	if overdue <= p.LateGrace {
		return 0
	}
	return math.Ceil(overdue.Hours()) * hourly * p.LateMultiplier
}
