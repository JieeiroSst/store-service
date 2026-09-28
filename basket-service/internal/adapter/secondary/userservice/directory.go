package userservice

import (
	"context"
	"errors"
	"strconv"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type directory struct{ c *Client }

func NewDirectory(c *Client) port.UserDirectory { return directory{c: c} }

func (d directory) GetByID(ctx context.Context, id int) (*model.User, error) {
	u, err := d.c.GetUser(ctx, strconv.Itoa(id))
	if errors.Is(err, ErrUserNotFound) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &model.User{ID: u.ID, Username: u.Username, Email: u.Email, Name: u.Name, Active: u.Active}, nil
}
