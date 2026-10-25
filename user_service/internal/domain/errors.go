package domain

import "strconv"

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
	CodeAccountLocked       ErrorCode = "AUTH_ACCOUNT_LOCKED"
	CodeTooManyAttempts     ErrorCode = "AUTH_TOO_MANY_ATTEMPTS"
	CodeSSOFailed           ErrorCode = "AUTH_SSO_FAILED"
	CodeMFARequired         ErrorCode = "AUTH_MFA_REQUIRED"
	CodeMFAInvalidCode      ErrorCode = "AUTH_MFA_INVALID_CODE"
	CodeMFAChallengeInvalid ErrorCode = "AUTH_MFA_CHALLENGE_INVALID"
	CodeMFANotEnabled       ErrorCode = "MFA_NOT_ENABLED"
	CodeMFAAlreadyEnabled   ErrorCode = "MFA_ALREADY_ENABLED"
	CodeMFANotSetUp         ErrorCode = "MFA_NOT_SET_UP"

	CodeUsernameRequired ErrorCode = "USER_USERNAME_REQUIRED"
	CodeUsernameTaken    ErrorCode = "USER_USERNAME_TAKEN"
	CodeInvalidEmail     ErrorCode = "USER_INVALID_EMAIL"
	CodeWeakPassword     ErrorCode = "USER_WEAK_PASSWORD"
	CodeUserNotFound     ErrorCode = "USER_NOT_FOUND"

	CodeSessionNotFound ErrorCode = "SESSION_NOT_FOUND"

	CodeRoleNotFound   ErrorCode = "ROLE_NOT_FOUND"
	CodeRoleNotDefined ErrorCode = "ROLE_NOT_DEFINED"
)

type Error struct {
	Code     ErrorCode
	Message  string
	Metadata map[string]string
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
	ErrAccountLocked       = newError(CodeAccountLocked, "account is locked")
	ErrTooManyAttempts     = newError(CodeTooManyAttempts, "too many failed login attempts, please try again later")
	ErrSSOStateMismatch    = newError(CodeSSOFailed, "sso state is missing or does not match, please login again")
	ErrSSOExchangeFailed   = newError(CodeSSOFailed, "sso authorization code could not be verified")
	ErrSSODenied           = newError(CodeSSOFailed, "sso login was denied by the identity provider")
	ErrSSOEmailUnverified  = newError(CodeSSOFailed, "identity provider did not return a verified email")
	ErrSSOIDTokenInvalid   = newError(CodeSSOFailed, "id_token is invalid or expired")
	ErrSSOTokenUnsupported = newError(CodeNotImplemented, "token login is not supported by this provider, use the redirect flow")
	ErrSSONoIdentity       = newError(CodeSSOFailed, "identity provider did not return a user id")
	ErrMFAInvalidCode      = newError(CodeMFAInvalidCode, "authentication code is invalid or already used")
	ErrMFAChallenge        = newError(CodeMFAChallengeInvalid, "two-factor challenge is invalid or expired, please login again")
	ErrMFANotEnabled       = newError(CodeMFANotEnabled, "two-factor authentication is not enabled")
	ErrMFAAlreadyEnabled   = newError(CodeMFAAlreadyEnabled, "two-factor authentication is already enabled")
	ErrMFANotSetUp         = newError(CodeMFANotSetUp, "two-factor authentication has not been set up, call setup first")

	ErrUsernameRequired = newError(CodeUsernameRequired, "username is required")
	ErrUserExist        = newError(CodeUsernameTaken, "username already exists")
	ErrEmailFailed      = newError(CodeInvalidEmail, "email does not satisfy the condition")
	ErrPasswordFailed   = newError(CodeWeakPassword, "password does not satisfy the condition")
	ErrIPFailed         = newError(CodeInvalidRequest, "IP does not satisfy the condition")
	ErrUserNotExist     = newError(CodeUserNotFound, "user does not exist")

	ErrSessionNotFound = newError(CodeSessionNotFound, "session not found")

	ErrRoleNotFound       = newError(CodeRoleNotFound, "role not found")
	ErrRoleNotInAuthorize = newError(CodeRoleNotDefined, "role is not defined in authorize service")

	ErrHashPasswordFailed = newError(CodeInternal, "hash password failed")
	ErrLockAccountFailed  = newError(CodeInternal, "lock account failed")
	ErrAssignRoleFailed   = newError(CodeServiceUnavailable, "assign role in authorize service failed")
)

func NewMFARequired(token string, expiresIn int64) *Error {
	return &Error{
		Code:    CodeMFARequired,
		Message: "two-factor authentication code required",
		Metadata: map[string]string{
			"mfa_token":  token,
			"mfa_method": MFAMethodTOTP,
			"expires_in": strconv.FormatInt(expiresIn, 10),
		},
	}
}

const UserCacheKey = "user_id_%d"
