package model

type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Driver struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	VehicleType string `json:"vehicle_type"`
	IsActive    bool   `json:"is_active"`
}

type Restaurant struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	IsActive       bool    `json:"is_active"`
	CommissionRate float64 `json:"commission_rate"`
}

type MenuItem struct {
	ID           string  `json:"id"`
	RestaurantID string  `json:"restaurant_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	IsActive     bool    `json:"is_active"`
}

type PaymentAuthorization struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}
