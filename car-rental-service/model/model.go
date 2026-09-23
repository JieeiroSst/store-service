package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id" db:"user_id"`
	Email          string    `json:"email" db:"email"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	FirstName      string    `json:"first_name" db:"first_name"`
	LastName       string    `json:"last_name" db:"last_name"`
	PhoneNumber    string    `json:"phone_number" db:"phone_number"`
	Address        string    `json:"address" db:"address"`
	UserType       UserType  `json:"user_type" db:"user_type"`
	DrivingLicense string    `json:"driving_license" db:"driving_license"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type UserType string

const (
	UserTypeCustomer UserType = "customer"
	UserTypeStaff    UserType = "staff"
	UserTypeAdmin    UserType = "admin"
)

type UserDocument struct {
	ID                 uuid.UUID          `json:"id" db:"document_id"`
	UserID             uuid.UUID          `json:"user_id" db:"user_id"`
	DocumentType       DocumentType       `json:"document_type" db:"document_type"`
	DocumentNumber     string             `json:"document_number" db:"document_number"`
	ExpiryDate         time.Time          `json:"expiry_date" db:"expiry_date"`
	DocumentURL        string             `json:"document_url" db:"document_url"`
	VerificationStatus VerificationStatus `json:"verification_status" db:"verification_status"`
	VerifiedBy         *uuid.UUID         `json:"verified_by" db:"verified_by"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at"`
}

type DocumentType string

const (
	DocumentTypeDrivingLicense DocumentType = "driving_license"
	DocumentTypeIDProof        DocumentType = "id_proof"
	DocumentTypePassport       DocumentType = "passport"
	DocumentTypeOther          DocumentType = "other"
)

type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusRejected VerificationStatus = "rejected"
)

type VehicleCategory struct {
	ID          uuid.UUID   `json:"id" db:"category_id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	VehicleType VehicleType `json:"vehicle_type" db:"vehicle_type"`
}

type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeBicycle    VehicleType = "bicycle"
)

type Vehicle struct {
	ID                 uuid.UUID          `json:"id" db:"vehicle_id"`
	CategoryID         uuid.UUID          `json:"category_id" db:"category_id"`
	RegistrationNumber string             `json:"registration_number" db:"registration_number"`
	Make               string             `json:"make" db:"make"`
	Model              string             `json:"model" db:"model"`
	Year               int                `json:"year" db:"year"`
	Color              string             `json:"color" db:"color"`
	Mileage            float64            `json:"mileage" db:"mileage"`
	Status             VehicleStatus      `json:"status" db:"status"`
	HourlyRate         float64            `json:"hourly_rate" db:"hourly_rate"`
	DailyRate          float64            `json:"daily_rate" db:"daily_rate"`
	LocationID         uuid.UUID          `json:"location_id" db:"location_id"`
	Features           map[string]string  `json:"features" db:"features"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at"`
	Category           *VehicleCategory   `json:"category,omitempty" db:"-"`
	Location           *Location          `json:"location,omitempty" db:"-"`
	CarDetails         *CarDetails        `json:"car_details,omitempty" db:"-"`
	MotorcycleDetails  *MotorcycleDetails `json:"motorcycle_details,omitempty" db:"-"`
	BicycleDetails     *BicycleDetails    `json:"bicycle_details,omitempty" db:"-"`
}

type VehicleStatus string

const (
	VehicleStatusAvailable   VehicleStatus = "available"
	VehicleStatusRented      VehicleStatus = "rented"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
	VehicleStatusRetired     VehicleStatus = "retired"
)

type CarDetails struct {
	ID              uuid.UUID `json:"-" db:"car_detail_id"`
	VehicleID       uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	FuelType        string    `json:"fuel_type" db:"fuel_type"`
	Transmission    string    `json:"transmission" db:"transmission"`
	SeatingCapacity int       `json:"seating_capacity" db:"seating_capacity"`
	TrunkCapacity   float64   `json:"trunk_capacity" db:"trunk_capacity"`
	AirConditioning bool      `json:"air_conditioning" db:"air_conditioning"`
}

type MotorcycleDetails struct {
	ID             uuid.UUID `json:"-" db:"motorcycle_detail_id"`
	VehicleID      uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	EngineCapacity int       `json:"engine_capacity" db:"engine_capacity"`
	MotorcycleType string    `json:"motorcycle_type" db:"motorcycle_type"`
	HelmetIncluded bool      `json:"helmet_included" db:"helmet_included"`
}

type BicycleDetails struct {
	ID          uuid.UUID `json:"-" db:"bicycle_detail_id"`
	VehicleID   uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	BicycleType string    `json:"bicycle_type" db:"bicycle_type"`
	FrameSize   string    `json:"frame_size" db:"frame_size"`
	GearCount   int       `json:"gear_count" db:"gear_count"`
	HasBasket   bool      `json:"has_basket" db:"has_basket"`
}

type Location struct {
	ID           uuid.UUID         `json:"id" db:"location_id"`
	Name         string            `json:"name" db:"name"`
	Address      string            `json:"address" db:"address"`
	City         string            `json:"city" db:"city"`
	State        string            `json:"state" db:"state"`
	Country      string            `json:"country" db:"country"`
	ZipCode      string            `json:"zip_code" db:"zip_code"`
	Latitude     float64           `json:"latitude" db:"latitude"`
	Longitude    float64           `json:"longitude" db:"longitude"`
	ContactPhone string            `json:"contact_phone" db:"contact_phone"`
	OpeningHours map[string]string `json:"opening_hours" db:"opening_hours"`
}

type Reservation struct {
	ID               uuid.UUID         `json:"id" db:"reservation_id"`
	UserID           uuid.UUID         `json:"user_id" db:"user_id"`
	VehicleID        uuid.UUID         `json:"vehicle_id" db:"vehicle_id"`
	PickupLocationID uuid.UUID         `json:"pickup_location_id" db:"pickup_location_id"`
	ReturnLocationID uuid.UUID         `json:"return_location_id" db:"return_location_id"`
	StartTime        time.Time         `json:"start_time" db:"start_time"`
	EndTime          time.Time         `json:"end_time" db:"end_time"`
	Status           ReservationStatus `json:"status" db:"status"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at" db:"updated_at"`
	User             *User             `json:"user,omitempty" db:"-"`
	Vehicle          *Vehicle          `json:"vehicle,omitempty" db:"-"`
	PickupLocation   *Location         `json:"pickup_location,omitempty" db:"-"`
	ReturnLocation   *Location         `json:"return_location,omitempty" db:"-"`
}

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusCompleted ReservationStatus = "completed"
)

type Rental struct {
	ID               uuid.UUID     `json:"id" db:"rental_id"`
	ReservationID    *uuid.UUID    `json:"reservation_id" db:"reservation_id"`
	VehicleID        uuid.UUID     `json:"vehicle_id" db:"vehicle_id"`
	UserID           uuid.UUID     `json:"user_id" db:"user_id"`
	PickupTime       time.Time     `json:"pickup_time" db:"pickup_time"`
	ActualReturnTime *time.Time    `json:"actual_return_time,omitempty" db:"actual_return_time"`
	PickupLocationID uuid.UUID     `json:"pickup_location_id" db:"pickup_location_id"`
	ReturnLocationID *uuid.UUID    `json:"return_location_id,omitempty" db:"return_location_id"`
	PickupMileage    float64       `json:"pickup_mileage" db:"pickup_mileage"`
	ReturnMileage    *float64      `json:"return_mileage,omitempty" db:"return_mileage"`
	Status           RentalStatus  `json:"status" db:"status"`
	BaseFee          float64       `json:"base_fee" db:"base_fee"`
	AdditionalFees   float64       `json:"additional_fees" db:"additional_fees"`
	PaymentStatus    PaymentStatus `json:"payment_status" db:"payment_status"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
	Reservation      *Reservation  `json:"reservation,omitempty" db:"-"`
	User             *User         `json:"user,omitempty" db:"-"`
	Vehicle          *Vehicle      `json:"vehicle,omitempty" db:"-"`
	PickupLocation   *Location     `json:"pickup_location,omitempty" db:"-"`
	ReturnLocation   *Location     `json:"return_location,omitempty" db:"-"`
	Payments         []*Payment    `json:"payments,omitempty" db:"-"`
}

type RentalStatus string

const (
	RentalStatusActive    RentalStatus = "active"
	RentalStatusCompleted RentalStatus = "completed"
	RentalStatusOverdue   RentalStatus = "overdue"
)

type Payment struct {
	ID            uuid.UUID     `json:"id" db:"payment_id"`
	RentalID      uuid.UUID     `json:"rental_id" db:"rental_id"`
	UserID        uuid.UUID     `json:"user_id" db:"user_id"`
	Amount        float64       `json:"amount" db:"amount"`
	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method"`
	TransactionID string        `json:"transaction_id" db:"transaction_id"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	PaymentDate   time.Time     `json:"payment_date" db:"payment_date"`
	Notes         string        `json:"notes" db:"notes"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	Rental        *Rental       `json:"rental,omitempty" db:"-"`
	User          *User         `json:"user,omitempty" db:"-"`
}

type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodDebitCard  PaymentMethod = "debit_card"
	PaymentMethodCash       PaymentMethod = "cash"
	PaymentMethodOnline     PaymentMethod = "online"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type MaintenanceRecord struct {
	ID                  uuid.UUID  `json:"id" db:"record_id"`
	VehicleID           uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	MaintenanceType     string     `json:"maintenance_type" db:"maintenance_type"`
	Description         string     `json:"description" db:"description"`
	Cost                float64    `json:"cost" db:"cost"`
	PerformedBy         string     `json:"performed_by" db:"performed_by"`
	MaintenanceDate     time.Time  `json:"maintenance_date" db:"maintenance_date"`
	NextMaintenanceDate *time.Time `json:"next_maintenance_date,omitempty" db:"next_maintenance_date"`
	Notes               string     `json:"notes" db:"notes"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	Vehicle             *Vehicle   `json:"vehicle,omitempty" db:"-"`
}

type Review struct {
	ID        uuid.UUID `json:"id" db:"review_id"`
	RentalID  uuid.UUID `json:"rental_id" db:"rental_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	VehicleID uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	Rating    int       `json:"rating" db:"rating"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	User      *User     `json:"user,omitempty" db:"-"`
	Vehicle   *Vehicle  `json:"vehicle,omitempty" db:"-"`
	Rental    *Rental   `json:"rental,omitempty" db:"-"`
}

type PaginationParams struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

type SearchVehicleParams struct {
	PaginationParams
	StartTime          *time.Time   `json:"start_time" form:"start_time"`
	EndTime            *time.Time   `json:"end_time" form:"end_time"`
	PickupLocationID   *uuid.UUID   `json:"pickup_location_id" form:"pickup_location_id"`
	ReturnLocationID   *uuid.UUID   `json:"return_location_id" form:"return_location_id"`
	VehicleType        *VehicleType `json:"vehicle_type" form:"vehicle_type"`
	CategoryID         *uuid.UUID   `json:"category_id" form:"category_id"`
	MinSeatingCapacity *int         `json:"min_seating_capacity" form:"min_seating_capacity"`
	MinEngineCapacity  *int         `json:"min_engine_capacity" form:"min_engine_capacity"`
	MaxDailyRate       *float64     `json:"max_daily_rate" form:"max_daily_rate"`
}
