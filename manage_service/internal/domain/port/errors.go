package port

import "errors"

var ErrAdminNotConfigured = errors.New("keycloak admin credentials are not configured")
