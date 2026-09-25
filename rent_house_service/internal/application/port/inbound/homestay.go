package inbound

import (
	"context"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type HomestayInput struct {
	Name        string
	Description string
	Type        int
	Status      domain.HomestayStatus
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
	WalletID    *string
}

type SetAvailabilityCommand struct {
	HomestayID int64
	From       time.Time // inclusive
	To         time.Time // exclusive
	Price      string    // decimal, per night
	Status     domain.AvailabilityStatus
}

type HomestayUseCase interface {
	View(ctx context.Context, viewer *Principal, id int64) (domain.Homestay, error)
	List(ctx context.Context, filter domain.HomestayFilter) (domain.HomestayPage, error)
	Availability(ctx context.Context, homestayID int64, from, to time.Time) ([]domain.Slot, error)
	ListAmenities(ctx context.Context) ([]domain.Amenity, error)
	Rates(ctx context.Context, homestayID int64) ([]domain.Rate, error)
	Mine(ctx context.Context, actor Principal, limit, offset int) ([]domain.Homestay, error)
	ForReview(ctx context.Context, actor Principal, status domain.HomestayStatus, limit, offset int) ([]domain.Homestay, error)
	Approve(ctx context.Context, actor Principal, id int64) (domain.Homestay, error)
	Reject(ctx context.Context, actor Principal, id int64, reason string) (domain.Homestay, error)
	Create(ctx context.Context, actor Principal, in HomestayInput) (domain.Homestay, error)
	Update(ctx context.Context, actor Principal, id int64, in HomestayInput) (domain.Homestay, error)
	Deactivate(ctx context.Context, actor Principal, id int64) error
	SetAvailability(ctx context.Context, actor Principal, cmd SetAvailabilityCommand) error
	CreateAmenity(ctx context.Context, actor Principal, name, icon string) (domain.Amenity, error)
	SetRate(ctx context.Context, actor Principal, r domain.Rate) error
	DeleteRate(ctx context.Context, actor Principal, homestayID int64, model domain.RentalModel) error
}
