package merchant

import "errors"

var (
	ErrMerchantNotFound        = errors.New("merchant not found")
	ErrMerchantInactive        = errors.New("merchant is inactive")
	ErrUnsupportedProviderType = errors.New("unsupported provider type")
	ErrInvalidMerchant         = errors.New("invalid merchant")
	ErrVersionConflict         = errors.New("merchant version conflict")
)
