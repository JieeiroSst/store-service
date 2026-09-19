package port

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrNoSpotAvailable      = errors.New("no compatible parking spot available")
	ErrTicketAlreadyClosed  = errors.New("ticket already checked out")
	ErrInvalidVehicleType   = errors.New("invalid vehicle type")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
)
