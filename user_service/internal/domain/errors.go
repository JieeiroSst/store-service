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
)

// UserCacheKey is the fmt.Sprintf pattern used to key a cached user by id.
const UserCacheKey = "user_id_%d"
