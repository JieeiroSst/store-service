package domain

import "time"

type HomestayStatus int16

const (
	HomestayActive   HomestayStatus = 1 // published
	HomestayInactive HomestayStatus = 2 // paused by its owner
	HomestayPending  HomestayStatus = 3 // registered, waiting for the platform admin
	HomestayRejected HomestayStatus = 4 // the admin turned it down; see ReviewNote
)

// PropertyType is what kind of place is being rented.
type PropertyType int

const (
	PropertyHomestay PropertyType = 1
	PropertyHotel    PropertyType = 2
	PropertyHouse    PropertyType = 3 // a house or flat for rent
)

func (t PropertyType) Valid() bool { return t >= PropertyHomestay && t <= PropertyHouse }

func (t PropertyType) Name() string {
	switch t {
	case PropertyHomestay:
		return "homestay"
	case PropertyHotel:
		return "hotel"
	case PropertyHouse:
		return "house"
	}
	return ""
}

func (s HomestayStatus) Name() string {
	switch s {
	case HomestayActive:
		return "active"
	case HomestayInactive:
		return "inactive"
	case HomestayPending:
		return "pending"
	case HomestayRejected:
		return "rejected"
	}
	return ""
}

type Homestay struct {
	ID          int64
	Name        string
	Description string
	Type        int
	Status      HomestayStatus
	PhoneNumber string
	Address     string
	WardID      int
	DistrictID  int
	ProvinceID  int
	Images      []string
	Guests      int
	Bedrooms    int
	Bathrooms   int
	AmenityIDs  []int
	// HostID is the manager who created the homestay; their reviews feed their standing in recompense-service.
	HostID int64
	// WalletID is the wallet in payment-wallet-service that receives this homestay's payments.
	WalletID string
	// ReviewNote is why the admin rejected the listing.
	ReviewNote  string
	Rating      float64 // average of all reviews, 0 when none
	ReviewCount int
	Rates       []Rate
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AvailabilityStatus int16

const (
	SlotAvailable AvailabilityStatus = 1
	SlotBooked    AvailabilityStatus = 2
	SlotBlocked   AvailabilityStatus = 3
)

// Slot is one night of a homestay. Price is a decimal string to avoid float rounding.
type Slot struct {
	HomestayID int64
	Date       time.Time
	Price      string
	Status     AvailabilityStatus
}

type Amenity struct {
	ID   int
	Name string
	Icon string
}

type HomestayFilter struct {
	ProvinceID int
	DistrictID int
	WardID     int
	Type       int
	MinGuests  int
	// HostID keeps only this owner's homestays; Status keeps one status (0 = active only).
	HostID     int64
	Status     HomestayStatus
	Query      string // matches name or address
	AmenityIDs []int  // homestay must have all of them
	// MinPrice/MaxPrice bound the rate of Model (day when empty); decimal strings, empty = no bound.
	Model    RentalModel
	MinPrice string
	MaxPrice string
	// CheckIn/CheckOut keep only homestays whose every night in the range is open for booking.
	CheckIn  time.Time
	CheckOut time.Time
	// Sort is one of recommended (default), rating, price_asc, price_desc, newest.
	Sort string
	// After continues a search from the end of the previous page. Its Sort must match Sort.
	After *HomestayCursor
	// Only active homestays are returned unless IncludeInactive is set (admin).
	IncludeInactive bool
	Limit           int
	Offset          int
}

type HomestayCursor struct {
	Sort   string
	ID     int64
	Rating float64 // rating sort
	Count  int     // rating and recommended sorts: number of reviews
	Price  string  // price sorts: the rate that was sorted on, "" when the homestay has none
	Score  float64 // recommended sort: rating and host tier combined
}

// HomestayPage is one page of a search. Next is nil on the last page.
type HomestayPage struct {
	Items []Homestay
	Next  *HomestayCursor
}
