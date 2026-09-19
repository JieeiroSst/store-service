package model

const (
	StatusBusy = 0
	StatusFree = 1
)

type Delivery struct {
	ShipID     int     `json:"ship_id,omitempty"`
	Name       string  `json:"name,omitempty"`
	Address    string  `json:"address,omitempty"`
	KitchenID  int     `json:"kitchen_id,omitempty"`
	Status     int     `json:"status,omitempty"`
	DistanceKm float64 `json:"distance_km,omitempty" form:"distance_km"`
	// Fee and ETAMinutes are computed server-side from DistanceKm (see
	// application.estimateFeeAndETA) — never trusted from the caller.
	Fee        float64 `json:"fee,omitempty"`
	ETAMinutes int     `json:"eta_minutes,omitempty"`
}
