package model

type VehicleType string

const (
	Motorcycle      VehicleType = "MOTORCYCLE"
	Car             VehicleType = "CAR"
	Truck           VehicleType = "TRUCK"
	Bus             VehicleType = "BUS"
	Bicycle         VehicleType = "BICYCLE"
	ElectricVehicle VehicleType = "ELECTRIC_VEHICLE"
)

func (v VehicleType) Valid() bool {
	switch v {
	case Motorcycle, Car, Truck, Bus, Bicycle, ElectricVehicle:
		return true
	default:
		return false
	}
}

type Vehicle struct {
	LicensePlate string      `gorm:"column:license_plate;primaryKey"`
	Type         VehicleType `gorm:"column:type"`
}

func (Vehicle) TableName() string { return "vehicles" }
