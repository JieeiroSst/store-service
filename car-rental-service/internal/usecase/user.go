package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/JIeeiroSst/car-rental-service/model"
)

type UserDirectory interface {
	GetUser(ctx context.Context, id int64) (*model.User, error)
}

func (u *Usecase) GetUser(ctx context.Context, id string) (*model.User, error) {
	uid, err := parseUserID("user_id", id)
	if err != nil {
		return nil, err
	}
	user, err := u.users.GetUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	profile, err := u.repos.Profiles.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	user.DrivingLicense = profile.DrivingLicense
	return user, nil
}

type UpdateUserInput struct {
	UserID                               string
	Email, FirstName, LastName           string
	PhoneNumber, Address, DrivingLicense string
}

func (u *Usecase) UpdateUser(ctx context.Context, in UpdateUserInput) (*model.User, error) {
	for _, v := range []string{in.Email, in.FirstName, in.LastName, in.PhoneNumber, in.Address} {
		if strings.TrimSpace(v) != "" {
			return nil, precondition("email, name, phone and address are managed by user-service")
		}
	}
	user, err := u.GetUser(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if license := strings.TrimSpace(in.DrivingLicense); license != "" {
		now := time.Now()
		if err := u.repos.Profiles.Upsert(ctx, &model.CustomerProfile{
			UserID: user.ID, DrivingLicense: license, CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return nil, err
		}
		user.DrivingLicense = license
	}
	return user, nil
}
