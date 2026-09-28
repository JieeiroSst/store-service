package userservice

import (
	"context"
	"errors"
	"strconv"

	"github.com/JIeeiroSst/car-rental-service/model"
)

type Directory struct{ C *Client }

func (d Directory) GetUser(ctx context.Context, id int64) (*model.User, error) {
	u, err := d.C.GetUser(ctx, strconv.FormatInt(id, 10))
	if errors.Is(err, ErrUserNotFound) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &model.User{
		ID: id, Email: u.Email, FirstName: u.Name, PhoneNumber: u.Phone,
		Address: u.Address, UserType: model.UserTypeCustomer,
	}, nil
}
