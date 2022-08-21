package model

import "time"

type Transaction struct {
	TransactionId int
	UserId        int
	CartId        int
	StartDate     time.Time
	Duration      int
	PaymentId     int
	Status        Status
	Payment       Payment
	Message       Message
}

type Status string

var (
	Pending   Status = "Pending"
	Completed        = "Completed"
	Destroy          = "Destroy"
)
