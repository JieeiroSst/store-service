package common

import "errors"

var (
	PasswordFailed = errors.New("password does not satisfy the condition")
	EmailFailed    = errors.New("email does not satisfy the condition")
	IPFailed       = errors.New("IP does not satisfy the condition")

	HashPasswordFailed = errors.New("password failed")
	UserAlready        = errors.New("user already exists")

	FailedToken = errors.New("Missing Authentication Token")

	FailedTokenUsername = errors.New("Missing Authentication Username Token")

	NotFound = errors.New("Not Found")

	LockAccountFailed = errors.New("lock account failed")

	UserNotExist = errors.New("user does not exist")

	UserExist = errors.New("user does exist")
)
