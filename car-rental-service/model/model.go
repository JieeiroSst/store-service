package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id" db:"user_id" gorm:"column:user_id;primaryKey"`
	Email          string    `json:"email" db:"email" gorm:"column:email"`
	PasswordHash   string    `json:"-" db:"password_hash" gorm:"column:password_hash"`
	FirstName      string    `json:"first_name" db:"first_name" gorm:"column:first_name"`
	LastName       string    `json:"last_name" db:"last_name" gorm:"column:last_name"`
	PhoneNumber    string    `json:"phone_number" db:"phone_number" gorm:"column:phone_number"`
	Address        string    `json:"address" db:"address" gorm:"column:address"`
	UserType       UserType  `json:"user_type" db:"user_type" gorm:"column:user_type"`
	DrivingLicense string    `json:"driving_license" db:"driving_license" gorm:"column:driving_license"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at" gorm:"column:updated_at"`
}

type UserType string

const (
	UserTypeCustomer UserType = "customer"
	UserTypeStaff    UserType = "staff"
	UserTypeAdmin    UserType = "admin"
)

type UserDocument struct {
	ID                 uuid.UUID          `json:"id" db:"document_id" gorm:"column:document_id;primaryKey"`
	UserID             uuid.UUID          `json:"user_id" db:"user_id" gorm:"column:user_id"`
	DocumentType       DocumentType       `json:"document_type" db:"document_type" gorm:"column:document_type"`
	DocumentNumber     string             `json:"document_number" db:"document_number" gorm:"column:document_number"`
	ExpiryDate         time.Time          `json:"expiry_date" db:"expiry_date" gorm:"column:expiry_date"`
	DocumentURL        string             `json:"document_url" db:"document_url" gorm:"column:document_url"`
	VerificationStatus VerificationStatus `json:"verification_status" db:"verification_status" gorm:"column:verification_status"`
	VerifiedBy         *uuid.UUID         `json:"verified_by" db:"verified_by" gorm:"column:verified_by"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at" gorm:"column:created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at" gorm:"column:updated_at"`
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
	ID          uuid.UUID   `json:"id" db:"category_id" gorm:"column:category_id;primaryKey"`
	Name        string      `json:"name" db:"name" gorm:"column:name"`
	Description string      `json:"description" db:"description" gorm:"column:description"`
	VehicleType VehicleType `json:"vehicle_type" db:"vehicle_type" gorm:"column:vehicle_type"`
}

type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeBicycle    VehicleType = "bicycle"
)

type Vehicle struct {
	ID                 uuid.UUID          `json:"id" db:"vehicle_id" gorm:"column:vehicle_id;primaryKey"`
	CategoryID         uuid.UUID          `json:"category_id" db:"category_id" gorm:"column:category_id"`
	RegistrationNumber string             `json:"registration_number" db:"registration_number" gorm:"column:registration_number"`
	Make               string             `json:"make" db:"make" gorm:"column:make"`
	Model              string             `json:"model" db:"model" gorm:"column:model"`
	Year               int                `json:"year" db:"year" gorm:"column:year"`
	Color              string             `json:"color" db:"color" gorm:"column:color"`
	Mileage            float64            `json:"mileage" db:"mileage" gorm:"column:mileage"`
	Status             VehicleStatus      `json:"status" db:"status" gorm:"column:status"`
	HourlyRate         float64            `json:"hourly_rate" db:"hourly_rate" gorm:"column:hourly_rate"`
	DailyRate          float64            `json:"daily_rate" db:"daily_rate" gorm:"column:daily_rate"`
	LocationID         uuid.UUID          `json:"location_id" db:"location_id" gorm:"column:location_id"`
	Features           map[string]string  `json:"features" db:"features" gorm:"column:features;serializer:json"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at" gorm:"column:created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at" gorm:"column:updated_at"`
	Category           *VehicleCategory   `json:"category,omitempty" gorm:"-"`
	Location           *Location          `json:"location,omitempty" gorm:"-"`
	CarDetails         *CarDetails        `json:"car_details,omitempty" gorm:"-"`
	MotorcycleDetails  *MotorcycleDetails `json:"motorcycle_details,omitempty" gorm:"-"`
	BicycleDetails     *BicycleDetails    `json:"bicycle_details,omitempty" gorm:"-"`
}

type VehicleStatus string

const (
	VehicleStatusAvailable   VehicleStatus = "available"
	VehicleStatusRented      VehicleStatus = "rented"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
	VehicleStatusRetired     VehicleStatus = "retired"
)

type CarDetails struct {
	ID              uuid.UUID `json:"-" db:"car_detail_id" gorm:"column:car_detail_id;primaryKey"`
	VehicleID       uuid.UUID `json:"vehicle_id" db:"vehicle_id" gorm:"column:vehicle_id"`
	FuelType        string    `json:"fuel_type" db:"fuel_type" gorm:"column:fuel_type"`
	Transmission    string    `json:"transmission" db:"transmission" gorm:"column:transmission"`
	SeatingCapacity int       `json:"seating_capacity" db:"seating_capacity" gorm:"column:seating_capacity"`
	TrunkCapacity   float64   `json:"trunk_capacity" db:"trunk_capacity" gorm:"column:trunk_capacity"`
	AirConditioning bool      `json:"air_conditioning" db:"air_conditioning" gorm:"column:air_conditioning"`
}

type MotorcycleDetails struct {
	ID             uuid.UUID `json:"-" db:"motorcycle_detail_id" gorm:"column:motorcycle_detail_id;primaryKey"`
	VehicleID      uuid.UUID `json:"vehicle_id" db:"vehicle_id" gorm:"column:vehicle_id"`
	EngineCapacity int       `json:"engine_capacity" db:"engine_capacity" gorm:"column:engine_capacity"`
	MotorcycleType string    `json:"motorcycle_type" db:"motorcycle_type" gorm:"column:motorcycle_type"`
	HelmetIncluded bool      `json:"helmet_included" db:"helmet_included" gorm:"column:helmet_included"`
}

type BicycleDetails struct {
	ID          uuid.UUID `json:"-" db:"bicycle_detail_id" gorm:"column:bicycle_detail_id;primaryKey"`
	VehicleID   uuid.UUID `json:"vehicle_id" db:"vehicle_id" gorm:"column:vehicle_id"`
	BicycleType string    `json:"bicycle_type" db:"bicycle_type" gorm:"column:bicycle_type"`
	FrameSize   string    `json:"frame_size" db:"frame_size" gorm:"column:frame_size"`
	GearCount   int       `json:"gear_count" db:"gear_count" gorm:"column:gear_count"`
	HasBasket   bool      `json:"has_basket" db:"has_basket" gorm:"column:has_basket"`
}

type Location struct {
	ID           uuid.UUID         `json:"id" db:"location_id" gorm:"column:location_id;primaryKey"`
	Name         string            `json:"name" db:"name" gorm:"column:name"`
	Address      string            `json:"address" db:"address" gorm:"column:address"`
	City         string            `json:"city" db:"city" gorm:"column:city"`
	State        string            `json:"state" db:"state" gorm:"column:state"`
	Country      string            `json:"country" db:"country" gorm:"column:country"`
	ZipCode      string            `json:"zip_code" db:"zip_code" gorm:"column:zip_code"`
	Latitude     float64           `json:"latitude" db:"latitude" gorm:"column:latitude"`
	Longitude    float64           `json:"longitude" db:"longitude" gorm:"column:longitude"`
	ContactPhone string            `json:"contact_phone" db:"contact_phone" gorm:"column:contact_phone"`
	OpeningHours map[string]string `json:"opening_hours" db:"opening_hours" gorm:"column:opening_hours;serializer:json"`
}

type Reservation struct {
	ID                 uuid.UUID         `json:"id" db:"reservation_id" gorm:"column:reservation_id;primaryKey"`
	UserID             uuid.UUID         `json:"user_id" db:"user_id" gorm:"column:user_id"`
	VehicleID          uuid.UUID         `json:"vehicle_id" db:"vehicle_id" gorm:"column:vehicle_id"`
	PickupLocationID   uuid.UUID         `json:"pickup_location_id" db:"pickup_location_id" gorm:"column:pickup_location_id"`
	ReturnLocationID   uuid.UUID         `json:"return_location_id" db:"return_location_id" gorm:"column:return_location_id"`
	StartTime          time.Time         `json:"start_time" db:"start_time" gorm:"column:start_time"`
	EndTime            time.Time         `json:"end_time" db:"end_time" gorm:"column:end_time"`
	Status             ReservationStatus `json:"status" db:"status" gorm:"column:status"`
	CancellationReason string            `json:"cancellation_reason,omitempty" db:"cancellation_reason" gorm:"column:cancellation_reason"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at" gorm:"column:created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at" gorm:"column:updated_at"`
	User               *User             `json:"user,omitempty" gorm:"-"`
	Vehicle            *Vehicle          `json:"vehicle,omitempty" gorm:"-"`
	PickupLocation     *Location         `json:"pickup_location,omitempty" gorm:"-"`
	ReturnLocation     *Location         `json:"return_location,omitempty" gorm:"-"`
}

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusCompleted ReservationStatus = "completed"
)

type Rental struct {
	ID               uuid.UUID     `json:"id" db:"rental_id" gorm:"column:rental_id;primaryKey"`
	ReservationID    *uuid.UUID    `json:"reservation_id" db:"reservation_id" gorm:"column:reservation_id"`
	VehicleID        uuid.UUID     `json:"vehicle_id" db:"vehicle_id" gorm:"column:vehicle_id"`
	UserID           uuid.UUID     `json:"user_id" db:"user_id" gorm:"column:user_id"`
	PickupTime       time.Time     `json:"pickup_time" db:"pickup_time" gorm:"column:pickup_time"`
	ActualReturnTime *time.Time    `json:"actual_return_time,omitempty" db:"actual_return_time" gorm:"column:actual_return_time"`
	PickupLocationID uuid.UUID     `json:"pickup_location_id" db:"pickup_location_id" gorm:"column:pickup_location_id"`
	ReturnLocationID *uuid.UUID    `json:"return_location_id,omitempty" db:"return_location_id" gorm:"column:return_location_id"`
	PickupMileage    float64       `json:"pickup_mileage" db:"pickup_mileage" gorm:"column:pickup_mileage"`
	ReturnMileage    *float64      `json:"return_mileage,omitempty" db:"return_mileage" gorm:"column:return_mileage"`
	Status           RentalStatus  `json:"status" db:"status" gorm:"column:status"`
	BaseFee          float64       `json:"base_fee" db:"base_fee" gorm:"column:base_fee"`
	AdditionalFees   float64       `json:"additional_fees" db:"additional_fees" gorm:"column:additional_fees"`
	Notes            string        `json:"notes,omitempty" db:"notes" gorm:"column:notes"`
	PaymentStatus    PaymentStatus `json:"payment_status" db:"payment_status" gorm:"column:payment_status"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at" gorm:"column:updated_at"`
	Reservation      *Reservation  `json:"reservation,omitempty" gorm:"-"`
	User             *User         `json:"user,omitempty" gorm:"-"`
	Vehicle          *Vehicle      `json:"vehicle,omitempty" gorm:"-"`
	PickupLocation   *Location     `json:"pickup_location,omitempty" gorm:"-"`
	ReturnLocation   *Location     `json:"return_location,omitempty" gorm:"-"`
	Payments         []*Payment    `json:"payments,omitempty" gorm:"-"`
}

type RentalStatus string

const (
	RentalStatusActive    RentalStatus = "active"
	RentalStatusCompleted RentalStatus = "completed"
	RentalStatusOverdue   RentalStatus = "overdue"
)

type Payment struct {
	ID            uuid.UUID     `json:"id" db:"payment_id" gorm:"column:payment_id;primaryKey"`
	RentalID      uuid.UUID     `json:"rental_id" db:"rental_id" gorm:"column:rental_id"`
	UserID        uuid.UUID     `json:"user_id" db:"user_id" gorm:"column:user_id"`
	Amount        float64       `json:"amount" db:"amount" gorm:"column:amount"`
	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method" gorm:"column:payment_method"`
	TransactionID string        `json:"transaction_id" db:"transaction_id" gorm:"column:transaction_id"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status" gorm:"column:payment_status"`
	PaymentDate   time.Time     `json:"payment_date" db:"payment_date" gorm:"column:payment_date"`
	Notes         string        `json:"notes" db:"notes" gorm:"column:notes"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at" gorm:"column:created_at"`
	Rental        *Rental       `json:"rental,omitempty" gorm:"-"`
	User          *User         `json:"user,omitempty" gorm:"-"`
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
	ID                  uuid.UUID  `json:"id" db:"record_id" gorm:"column:record_id;primaryKey"`
	VehicleID           uuid.UUID  `json:"vehicle_id" db:"vehicle_id" gorm:"column:vehicle_id"`
	MaintenanceType     string     `json:"maintenance_type" db:"maintenance_type" gorm:"column:maintenance_type"`
	Description         string     `json:"description" db:"description" gorm:"column:description"`
	Cost                float64    `json:"cost" db:"cost" gorm:"column:cost"`
	PerformedBy         string     `json:"performed_by" db:"performed_by" gorm:"column:performed_by"`
	MaintenanceDate     time.Time  `json:"maintenance_date" db:"maintenance_date" gorm:"column:maintenance_date"`
	NextMaintenanceDate *time.Time `json:"next_maintenance_date,omitempty" db:"next_maintenance_date" gorm:"column:next_maintenance_date"`
	Notes               string     `json:"notes" db:"notes" gorm:"column:notes"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at" gorm:"column:created_at"`
	Vehicle             *Vehicle   `json:"vehicle,omitempty" gorm:"-"`
}

type Review struct {
	ID        uuid.UUID `json:"id" db:"review_id" gorm:"column:review_id;primaryKey"`
	RentalID  uuid.UUID `json:"rental_id" db:"rental_id" gorm:"column:rental_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id" gorm:"column:user_id"`
	VehicleID uuid.UUID `json:"vehicle_id" db:"vehicle_id" gorm:"column:vehicle_id"`
	Rating    int       `json:"rating" db:"rating" gorm:"column:rating"`
	Comment   string    `json:"comment" db:"comment" gorm:"column:comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at" gorm:"column:created_at"`
	User      *User     `json:"user,omitempty" gorm:"-"`
	Vehicle   *Vehicle  `json:"vehicle,omitempty" gorm:"-"`
	Rental    *Rental   `json:"rental,omitempty" gorm:"-"`
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

func (User) TableName() string              { return "users" }
func (UserDocument) TableName() string      { return "user_documents" }
func (VehicleCategory) TableName() string   { return "vehicle_categories" }
func (Vehicle) TableName() string           { return "vehicles" }
func (CarDetails) TableName() string        { return "car_details" }
func (MotorcycleDetails) TableName() string { return "motorcycle_details" }
func (BicycleDetails) TableName() string    { return "bicycle_details" }
func (Location) TableName() string          { return "locations" }
func (Reservation) TableName() string       { return "reservations" }
func (Rental) TableName() string            { return "rentals" }
func (Payment) TableName() string           { return "payments" }
func (MaintenanceRecord) TableName() string { return "maintenance_records" }
func (Review) TableName() string            { return "reviews" }
