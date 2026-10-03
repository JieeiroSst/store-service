package grpcadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorResponse(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		viaGRPC    bool // pass through the interceptor first, as a handler error would
		wantCode   domain.ErrorCode
		wantStatus int
		wantMsg    string
	}{
		{"domain error", domain.ErrInvalidCredentials, true, domain.CodeInvalidCredentials, http.StatusUnauthorized, "invalid username or password"},
		{"wrapped domain error", fmt.Errorf("%w: boom", domain.ErrAssignRoleFailed), true, domain.CodeServiceUnavailable, http.StatusServiceUnavailable, domain.ErrAssignRoleFailed.Message},
		{"username taken", domain.ErrUserExist, true, domain.CodeUsernameTaken, http.StatusConflict, domain.ErrUserExist.Message},
		{"raw error hidden", errors.New("pq: connection refused"), true, domain.CodeInternal, http.StatusInternalServerError, "internal server error"},
		{"unimplemented rpc", status.Error(codes.Unimplemented, "method AddRole not implemented"), true, domain.CodeNotImplemented, http.StatusNotImplemented, "not implemented"},
		{"upstream unavailable", status.Error(codes.Unavailable, "authorize down"), true, domain.CodeServiceUnavailable, http.StatusServiceUnavailable, domain.ErrServiceUnavailable.Message},
		{"gateway bad body", status.Error(codes.InvalidArgument, "invalid character 'x'"), false, domain.CodeInvalidRequest, http.StatusBadRequest, "invalid request"},
		{"gateway unknown route", status.Error(codes.NotFound, "Not Found"), false, domain.CodeRouteNotFound, http.StatusNotFound, "route not found"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.err
			if c.viaGRPC {
				err = toStatus("/user.UserService/Test", err)
			}

			rec := httptest.NewRecorder()
			HTTPErrorHandler(context.Background(), nil, &runtime.JSONPb{}, rec, httptest.NewRequest(http.MethodPost, "/", nil), err)

			if rec.Code != c.wantStatus {
				t.Errorf("HTTP status = %d, want %d", rec.Code, c.wantStatus)
			}
			var got HTTPError
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode body %q: %v", rec.Body.String(), err)
			}
			want := HTTPError{Code: c.wantCode, Message: c.wantMsg, Status: c.wantStatus}
			if got != want {
				t.Errorf("body = %+v, want %+v", got, want)
			}
		})
	}
}
