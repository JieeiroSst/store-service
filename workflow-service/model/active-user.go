package model

import "time"

type ActiveUser struct {
	ID          string
	Key         string
	Value       string
	UserPeding  string
	UserApprove string
	UserReject  string
	Status      string
	CreateAt    time.Time
	UpdateAt    time.Time
	DeleteAt    time.Time
}
