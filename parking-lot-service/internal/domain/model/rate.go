package model

import "time"

type Rate struct {
	RateID        string      `gorm:"column:rate_id;primaryKey"`
	VehicleType   VehicleType `gorm:"column:vehicle_type"`
	HourlyRate    float64     `gorm:"column:hourly_rate"`
	DailyMaxRate  *float64    `gorm:"column:daily_max_rate"`
	EffectiveFrom time.Time   `gorm:"column:effective_from"`
}

func (Rate) TableName() string { return "rates" }

func (r Rate) Calculate(duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}

	hours := int(duration / time.Hour)
	if duration%time.Hour != 0 {
		hours++
	}

	fullDays := hours / 24
	remainderHours := hours % 24

	total := float64(fullDays) * r.dailyRate()
	remainderCost := float64(remainderHours) * r.HourlyRate
	if r.DailyMaxRate != nil && remainderCost > *r.DailyMaxRate {
		remainderCost = *r.DailyMaxRate
	}
	return total + remainderCost
}

func (r Rate) dailyRate() float64 {
	if r.DailyMaxRate != nil {
		return *r.DailyMaxRate
	}
	return r.HourlyRate * 24
}
