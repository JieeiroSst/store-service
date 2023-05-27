package model

import (
	"fmt"
	"time"
)

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

func (e STATUS) String() string {
	switch e {
	case PENDING:
		return "pending"
	case RESOLVED:
		return "resolved"
	default:
		return fmt.Sprintf("%v", string(e))
	}
}

func ParseStatus(s string) (c STATUS, err error) {
	status := map[STATUS]struct{}{
		PENDING:  {},
		RESOLVED: {},
	}
	cap := STATUS(s)
	_, ok := status[cap]
	if !ok {
		return c, fmt.Errorf(`cannot parse:[%s] as capability`, s)
	}
	return cap, nil
}
