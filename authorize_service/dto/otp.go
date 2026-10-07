package dto

import "time"

type OTP struct {
	OTP       string
	ExpiresAt time.Time
}
