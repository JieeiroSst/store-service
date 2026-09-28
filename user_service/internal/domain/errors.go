package domain

import "errors"

var (
	ErrPasswordFailed = errors.New("password does not satisfy the condition")
	ErrEmailFailed    = errors.New("email does not satisfy the condition")
	ErrIPFailed       = errors.New("IP does not satisfy the condition")

	ErrHashPasswordFailed = errors.New("password failed")
	ErrUserAlready        = errors.New("user already exists")

	ErrFailedToken = errors.New("Missing Authentication Token")

	ErrFailedTokenUsername = errors.New("Missing Authentication Username Token")

	ErrNotFound = errors.New("Not Found")

	ErrLockAccountFailed = errors.New("lock account failed")

	ErrUserNotExist = errors.New("user does not exist")

	ErrUserExist = errors.New("user does exist")

	ErrAssignRoleFailed = errors.New("assign role in authorize service failed")

	ErrRoleNotInAuthorize = errors.New("role is not defined in authorize service")
)

const UserCacheKey = "user_id_%d"
