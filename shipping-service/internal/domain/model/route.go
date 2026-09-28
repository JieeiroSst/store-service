package model

import (
	"sort"
	"time"
)

type Strategy string

const (
	StrategyCheapest Strategy = "CHEAPEST"
	StrategyFastest  Strategy = "FASTEST"
	StrategyBalanced Strategy = "BALANCED"
)

func (s Strategy) Valid() bool {
	switch s {
	case StrategyCheapest, StrategyFastest, StrategyBalanced:
		return true
	}
	return false
}

type RouteOption struct {
	WarehouseID        int64     `json:"warehouse_id"`
	WarehouseCode      string    `json:"warehouse_code"`
	CarrierShopID      int64     `json:"carrier_shop_id"`
	ServiceID          int       `json:"service_id"`
	ServiceTypeID      int       `json:"service_type_id"`
	ServiceName        string    `json:"service_name"`
	Fee                int64     `json:"fee"`
	ExpectedDeliveryAt time.Time `json:"expected_delivery_at"`
	Score              float64   `json:"score"`
}

type UnavailableRoute struct {
	WarehouseCode string `json:"warehouse_code"`
	ServiceID     int    `json:"service_id,omitempty"`
	Reason        string `json:"reason"`
}

const balancedFeeWeight = 0.5

func RankRoutes(options []RouteOption, strategy Strategy) []RouteOption {
	ranked := append([]RouteOption(nil), options...)
	if len(ranked) == 0 {
		return ranked
	}

	minFee, maxFee := ranked[0].Fee, ranked[0].Fee
	minETA, maxETA := ranked[0].ExpectedDeliveryAt, ranked[0].ExpectedDeliveryAt
	for _, o := range ranked[1:] {
		minFee, maxFee = min(minFee, o.Fee), max(maxFee, o.Fee)
		if o.ExpectedDeliveryAt.Before(minETA) {
			minETA = o.ExpectedDeliveryAt
		}
		if o.ExpectedDeliveryAt.After(maxETA) {
			maxETA = o.ExpectedDeliveryAt
		}
	}

	for i := range ranked {
		fee := normalize(float64(ranked[i].Fee-minFee), float64(maxFee-minFee))
		eta := normalize(float64(ranked[i].ExpectedDeliveryAt.Sub(minETA)), float64(maxETA.Sub(minETA)))
		switch strategy {
		case StrategyCheapest:
			ranked[i].Score = fee
		case StrategyFastest:
			ranked[i].Score = eta
		default:
			ranked[i].Score = balancedFeeWeight*fee + (1-balancedFeeWeight)*eta
		}
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		a, b := ranked[i], ranked[j]
		if a.Score != b.Score {
			return a.Score < b.Score
		}
		if strategy == StrategyFastest {
			if !a.ExpectedDeliveryAt.Equal(b.ExpectedDeliveryAt) {
				return a.ExpectedDeliveryAt.Before(b.ExpectedDeliveryAt)
			}
			return a.Fee < b.Fee
		}
		if a.Fee != b.Fee {
			return a.Fee < b.Fee
		}
		return a.ExpectedDeliveryAt.Before(b.ExpectedDeliveryAt)
	})
	return ranked
}

func normalize(v, span float64) float64 {
	if span <= 0 {
		return 0
	}
	return v / span
}
