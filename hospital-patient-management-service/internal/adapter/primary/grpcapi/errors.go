package grpcapi

import (
	"context"
	"errors"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toStatus maps domain errors to gRPC codes. Anything unexpected is logged and
// returned as an opaque Internal so database details never reach clients.
func toStatus(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	switch {
	case errors.Is(err, model.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, model.ErrPaymentFailed):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, model.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "unauthenticated")
	case errors.Is(err, model.ErrUpstream):
		logrus.WithError(err).Error("dependency unavailable")
		return status.Error(codes.Unavailable, "a dependent service is unavailable")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	}
	logrus.WithError(err).Error("request failed")
	return status.Error(codes.Internal, "internal error")
}
