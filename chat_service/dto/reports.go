package dto

import "time"

type Reports struct {
	ID             int
	UserId         int
	ParticipantsId int
	ReportType     string
	Notes          string
	Status         string
	CreatedAt      time.Time
}
