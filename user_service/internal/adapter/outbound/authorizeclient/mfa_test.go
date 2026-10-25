package authorizeclient

import (
	"errors"
	"testing"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func withReason(code codes.Code, domainName, reason string) error {
	st, _ := status.New(code, "x").WithDetails(&errdetails.ErrorInfo{Reason: reason, Domain: domainName})
	return st.Err()
}

func TestMFAErrorMapping(t *testing.T) {
	cases := []struct {
		in   error
		want error
	}{
		{withReason(codes.Unauthenticated, authorizeErrorDomain, "MFA_INVALID_CODE"), domain.ErrMFAInvalidCode},
		{withReason(codes.ResourceExhausted, authorizeErrorDomain, "MFA_TOO_MANY_ATTEMPTS"), domain.ErrTooManyAttempts},
		{withReason(codes.FailedPrecondition, authorizeErrorDomain, "MFA_ALREADY_ENABLED"), domain.ErrMFAAlreadyEnabled},
		{withReason(codes.FailedPrecondition, authorizeErrorDomain, "MFA_NOT_SET_UP"), domain.ErrMFANotSetUp},
		{withReason(codes.FailedPrecondition, authorizeErrorDomain, "MFA_NOT_ENABLED"), domain.ErrMFANotEnabled},
	}
	for _, c := range cases {
		if got := mfaError(c.in); !errors.Is(got, c.want) {
			t.Errorf("mfaError(%v) = %v, want %v", c.in, got, c.want)
		}
	}

	foreign := withReason(codes.Unauthenticated, "other-service", "MFA_INVALID_CODE")
	if got := mfaError(foreign); got != foreign {
		t.Errorf("reasons from other domains must pass through, got %v", got)
	}
	down := status.Error(codes.Unavailable, "connection refused")
	if got := mfaError(down); status.Code(got) != codes.Unavailable {
		t.Errorf("transport errors must pass through, got %v", got)
	}
	if mfaError(nil) != nil {
		t.Error("nil must stay nil")
	}
}
