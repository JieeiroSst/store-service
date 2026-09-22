package port

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found in user-service")
	ErrIdentityExists       = errors.New("citizen identity already submitted for this user")
	ErrIdentityNotFound     = errors.New("citizen identity not found")
	ErrFaceNotFound         = errors.New("face biometric not found")
	ErrNoFaceDetected       = errors.New("no face detected in image")
	ErrMRZNotReadable       = errors.New("MRZ could not be read with sufficient confidence")
	ErrVerificationNotFound = errors.New("verification not found")
)
