package model

import "time"

type STATUS string
const (
	PENDING  STATUS = "pending"
	RESOLVED STATUS = "resolved"
)

type Reports struct {
	ID             int       `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId         int       `json:"users_id,omitempty" bson:"users_id,omitempty"`
	ParticipantsId int       `json:"participants_id,omitempty" bson:"participants_id,omitempty"`
	ReportType     string    `json:"report_type,omitempty" bson:"report_type,omitempty"`
	Notes          string    `json:"notes,omitempty" bson:"notes,omitempty"`
	Status         STATUS    `json:"status,omitempty" bson:"status,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
}
