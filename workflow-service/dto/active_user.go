package dto

import (
	"time"

	"github.com/JIeeiroSst/workflow-service/model"
)

type ActiveUser struct {
	ID       string
	Key      string
	Value    string
	User     string
	Status   string
	CreateAt time.Time
	UpdateAt time.Time
	DeleteAt time.Time
}

func FormatActiveUser(user ActiveUser) model.ActiveUser {
	return model.ActiveUser{
		ID:          user.ID,
		Key:         user.Key,
		Value:       user.Value,
		UserPeding:  user.User,
		UserApprove: user.User,
		UserReject:  user.User,
		Status:      user.Status,
		CreateAt:    user.CreateAt,
		UpdateAt:    user.UpdateAt,
		DeleteAt:    user.DeleteAt,
	}
}
