package model

type SpotType string

const (
	MotorcycleSpot SpotType = "MOTORCYCLE_SPOT"
	CompactSpot    SpotType = "COMPACT_SPOT"
	LargeSpot      SpotType = "LARGE_SPOT"
	HandicapSpot   SpotType = "HANDICAP_SPOT"
	EVChargingSpot SpotType = "EV_CHARGING_SPOT"
)

func CompatibleSpotTypes(v VehicleType) []SpotType {
	switch v {
	case Motorcycle, Bicycle:
		return []SpotType{MotorcycleSpot}
	case Car:
		return []SpotType{CompactSpot, HandicapSpot}
	case Truck, Bus:
		return []SpotType{LargeSpot}
	case ElectricVehicle:
		return []SpotType{EVChargingSpot, CompactSpot}
	default:
		return nil
	}
}

type ParkingSpot struct {
	SpotID      string   `gorm:"column:spot_id;primaryKey"`
	FloorID     *string  `gorm:"column:floor_id"`
	Type        SpotType `gorm:"column:type"`
	IsAvailable bool     `gorm:"column:is_available"`
}

func (ParkingSpot) TableName() string { return "parking_spots" }

type SpotAvailability struct {
	Type      SpotType `json:"type"`
	Available int64    `json:"available"`
	Total     int64    `json:"total"`
}
