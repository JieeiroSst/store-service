package grpcadapter

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errorDomain = "user-service"

var grpcCodes = map[domain.ErrorCode]codes.Code{
	domain.CodeInvalidRequest:      codes.InvalidArgument,    // 400
	domain.CodeRouteNotFound:       codes.NotFound,           // 404
	domain.CodeNotImplemented:      codes.Unimplemented,      // 501
	domain.CodeServiceUnavailable:  codes.Unavailable,        // 503
	domain.CodeInternal:            codes.Internal,           // 500
	domain.CodeInvalidCredentials:  codes.Unauthenticated,    // 401
	domain.CodeTokenInvalid:        codes.Unauthenticated,    // 401
	domain.CodeRefreshTokenInvalid: codes.Unauthenticated,    // 401
	domain.CodeUsernameRequired:    codes.InvalidArgument,    // 400
	domain.CodeUsernameTaken:       codes.AlreadyExists,      // 409
	domain.CodeInvalidEmail:        codes.InvalidArgument,    // 400
	domain.CodeWeakPassword:        codes.InvalidArgument,    // 400
	domain.CodeUserNotFound:        codes.NotFound,           // 404
	domain.CodeRoleNotFound:        codes.NotFound,           // 404
	domain.CodeRoleNotDefined:      codes.FailedPrecondition, // 400
}

func UnaryErrorInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return resp, toStatus(info.FullMethod, err)
	}
	return resp, nil
}

func toStatus(method string, err error) error {
	var de *domain.Error
	if !errors.As(err, &de) {
		de = classify(err)
		log.Printf("%s: %v", method, err)
	} else if de.Code == domain.CodeInternal || de.Code == domain.CodeServiceUnavailable {
		log.Printf("%s: %v", method, err)
	}

	code, ok := grpcCodes[de.Code]
	if !ok {
		code = codes.Internal
	}
	st, detErr := status.New(code, de.Message).WithDetails(&errdetails.ErrorInfo{
		Reason: string(de.Code),
		Domain: errorDomain,
	})
	if detErr != nil {
		return status.Error(code, de.Message)
	}
	return st.Err()
}

func classify(err error) *domain.Error {
	switch status.Code(err) {
	case codes.Unimplemented:
		return domain.ErrNotImplemented
	case codes.Unavailable, codes.DeadlineExceeded:
		return domain.ErrServiceUnavailable
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return domain.ErrServiceUnavailable
	}
	return domain.ErrInternal
}

type HTTPError struct {
	Code    domain.ErrorCode `json:"code"`
	Message string           `json:"message"`
	Status  int              `json:"status"`
}

func HTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	st := status.Convert(err)
	httpStatus := runtime.HTTPStatusFromCode(st.Code())

	body := HTTPError{Message: st.Message(), Status: httpStatus}
	for _, d := range st.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok && info.GetDomain() == errorDomain {
			body.Code = domain.ErrorCode(info.GetReason())
			break
		}
	}
	if body.Code == "" {
		fallback := gatewayError(st.Code())
		body.Code, body.Message = fallback.Code, fallback.Message
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write error response: %v", err)
	}
}

func gatewayError(c codes.Code) *domain.Error {
	switch c {
	case codes.InvalidArgument:
		return domain.ErrInvalidRequest
	case codes.NotFound:
		return domain.ErrRouteNotFound
	case codes.Unimplemented:
		return domain.ErrNotImplemented
	case codes.Unavailable, codes.DeadlineExceeded:
		return domain.ErrServiceUnavailable
	}
	return domain.ErrInternal
}
