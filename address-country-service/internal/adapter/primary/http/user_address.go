package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/pkg/authn"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateUserAddress(c *gofr.Context) (interface{}, error) {
	p, err := authn.Require(c)
	if err != nil {
		return nil, err
	}
	var req model.UserAddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if !p.IsAdmin() {
		uid, err := principalUserID(p)
		if err != nil {
			return nil, err
		}
		req.UserID = uid
	}

	if err := h.userAddress.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetUserAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.ownedUserAddress(c, id)
}

func (h *Handler) ownedUserAddress(c *gofr.Context, id int) (*model.UserAddress, error) {
	p, err := authn.Require(c)
	if err != nil {
		return nil, err
	}
	addr, err := h.userAddress.Get(c, id)
	if err != nil {
		return nil, err
	}
	if !p.Owns(strconv.Itoa(addr.UserID)) {
		return nil, authn.ErrForbidden
	}
	return addr, nil
}

func (h *Handler) UpdateUserAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	existing, err := h.ownedUserAddress(c, id)
	if err != nil {
		return nil, err
	}
	var req model.UserAddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.ID = id
	if p, _ := authn.Require(c); !p.IsAdmin() {
		req.UserID = existing.UserID
	}

	if err := h.userAddress.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteUserAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if _, err := h.ownedUserAddress(c, id); err != nil {
		return nil, err
	}
	if err := h.userAddress.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListUserAddresses(c *gofr.Context) (interface{}, error) {
	p, err := authn.Require(c)
	if err != nil {
		return nil, err
	}
	all, err := h.userAddress.List(c)
	if err != nil || p.IsAdmin() {
		return all, err
	}
	own := make([]model.UserAddress, 0, len(all))
	for _, a := range all {
		if strconv.Itoa(a.UserID) == p.UserID {
			own = append(own, a)
		}
	}
	return own, nil
}
