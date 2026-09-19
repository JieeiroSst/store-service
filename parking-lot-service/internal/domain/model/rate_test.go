package model

import (
	"testing"
	"time"
)

func TestRateCalculate(t *testing.T) {
	dailyMax := 150000.0
	rate := Rate{HourlyRate: 20000, DailyMaxRate: &dailyMax}

	cases := []struct {
		name     string
		duration time.Duration
		want     float64
	}{
		{"zero duration", 0, 0},
		{"under an hour rounds up", 30 * time.Minute, 20000},
		{"exact hours", 2 * time.Hour, 40000},
		{"partial hour rounds up", 2*time.Hour + time.Minute, 60000},
		{"capped at daily max", 10 * time.Hour, 150000},
		{"multi-day stay", 26 * time.Hour, 150000 + 20000*2}, // 1 full day + 2 billed hours
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rate.Calculate(tc.duration); got != tc.want {
				t.Errorf("Calculate(%s) = %v, want %v", tc.duration, got, tc.want)
			}
		})
	}
}

func TestCompatibleSpotTypes(t *testing.T) {
	cases := []struct {
		vehicle VehicleType
		want    []SpotType
	}{
		{Motorcycle, []SpotType{MotorcycleSpot}},
		{Bicycle, []SpotType{MotorcycleSpot}},
		{Car, []SpotType{CompactSpot, HandicapSpot}},
		{Truck, []SpotType{LargeSpot}},
		{Bus, []SpotType{LargeSpot}},
		{ElectricVehicle, []SpotType{EVChargingSpot, CompactSpot}},
	}

	for _, tc := range cases {
		t.Run(string(tc.vehicle), func(t *testing.T) {
			got := CompatibleSpotTypes(tc.vehicle)
			if len(got) != len(tc.want) {
				t.Fatalf("CompatibleSpotTypes(%s) = %v, want %v", tc.vehicle, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("CompatibleSpotTypes(%s)[%d] = %v, want %v", tc.vehicle, i, got[i], tc.want[i])
				}
			}
		})
	}
}
