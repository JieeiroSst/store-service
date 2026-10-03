package domain

type ErrorCode string

const (
	CodeInvalidRequest     ErrorCode = "INVALID_REQUEST"
	CodeRouteNotFound      ErrorCode = "ROUTE_NOT_FOUND"
	CodeNotImplemented     ErrorCode = "NOT_IMPLEMENTED"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	CodeInternal           ErrorCode = "INTERNAL_ERROR"

	CodeInvalidCredentials  ErrorCode = "AUTH_INVALID_CREDENTIALS"
	CodeTokenInvalid        ErrorCode = "AUTH_TOKEN_INVALID"
	CodeRefreshTokenInvalid ErrorCode = "AUTH_REFRESH_TOKEN_INVALID"

	CodeUsernameRequired ErrorCode = "USER_USERNAME_REQUIRED"
	CodeUsernameTaken    ErrorCode = "USER_USERNAME_TAKEN"
	CodeInvalidEmail     ErrorCode = "USER_INVALID_EMAIL"
	CodeWeakPassword     ErrorCode = "USER_WEAK_PASSWORD"
	CodeUserNotFound     ErrorCode = "USER_NOT_FOUND"

	CodeRoleNotFound   ErrorCode = "ROLE_NOT_FOUND"
	CodeRoleNotDefined ErrorCode = "ROLE_NOT_DEFINED"
)

type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string { return e.Message }

func newError(code ErrorCode, msg string) *Error { return &Error{Code: code, Message: msg} }

var (
	ErrInvalidRequest     = newError(CodeInvalidRequest, "invalid request")
	ErrRouteNotFound      = newError(CodeRouteNotFound, "route not found")
	ErrNotImplemented     = newError(CodeNotImplemented, "not implemented")
	ErrServiceUnavailable = newError(CodeServiceUnavailable, "service temporarily unavailable, please try again")
	ErrInternal           = newError(CodeInternal, "internal server error")

	ErrInvalidCredentials  = newError(CodeInvalidCredentials, "invalid username or password")
	ErrFailedToken         = newError(CodeTokenInvalid, "missing or invalid access token")
	ErrFailedTokenUsername = newError(CodeTokenInvalid, "access token does not belong to this username")
	ErrRefreshTokenInvalid = newError(CodeRefreshTokenInvalid, "refresh token is invalid or expired, please login again")

	ErrUsernameRequired = newError(CodeUsernameRequired, "username is required")
	ErrUserExist        = newError(CodeUsernameTaken, "username already exists")
	ErrEmailFailed      = newError(CodeInvalidEmail, "email does not satisfy the condition")
	ErrPasswordFailed   = newError(CodeWeakPassword, "password does not satisfy the condition")
	ErrIPFailed         = newError(CodeInvalidRequest, "IP does not satisfy the condition")
	ErrUserNotExist     = newError(CodeUserNotFound, "user does not exist")

	ErrRoleNotFound       = newError(CodeRoleNotFound, "role not found")
	ErrRoleNotInAuthorize = newError(CodeRoleNotDefined, "role is not defined in authorize service")

	ErrHashPasswordFailed = newError(CodeInternal, "hash password failed")
	ErrLockAccountFailed  = newError(CodeInternal, "lock account failed")
	ErrAssignRoleFailed   = newError(CodeServiceUnavailable, "assign role in authorize service failed")
)

const UserCacheKey = "user_id_%d"
