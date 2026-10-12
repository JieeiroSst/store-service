package domain

import "errors"

var ErrAccountRejected = errors.New("account rejected")

type Account struct {
	Username string
	Password string
	Email    string
	Name     string
}
