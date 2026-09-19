package common

import "errors"

var (
	ErrNotFound           = errors.New("record not found")
	ErrDBFailed           = errors.New("database operation failed")
	ErrTableAlreadyBooked = errors.New("table already booked for that time")
	ErrUnknownFoodInOrder = errors.New("order references a food id kitchen-service does not know")
)
