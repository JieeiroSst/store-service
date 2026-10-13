package internalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeUsers struct {
	input.UserService
}

func (fakeUsers) FindUser(_ context.Context, req dto.FindUserRequest) (dto.FindUserResponse, error) {
	if req.UserId != 7 {
		return dto.FindUserResponse{}, domain.ErrUserNotExist
	}
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return dto.FindUserResponse{Total: 1, Users: []*dto.User{{
		Id: 7, Username: "quanluu", Email: "quanluu@gmail.com", Name: "Quan Luu", Phone: "0901234567",
		Address: "HCM", Sex: "male", Checked: true, CreateTime: timestamppb.New(created),
	}}}, nil
}

func serve(t *testing.T, tokens []string, path string, header ...string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	mux := runtime.NewServeMux()
	if err := New(fakeUsers{}, fakeSessions{}, tokens).Register(mux); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec, body
}

func TestGetUser(t *testing.T) {
	rec, body := serve(t, nil, "/internal/v1/users/7")
	if rec.Code != 200 || body["username"] != "quanluu" || body["active"] != true || body["create_time"] != "2026-01-02T03:04:05Z" {
		t.Fatalf("%d %v", rec.Code, body)
	}
	if _, ok := body["password"]; ok {
		t.Fatal("password must never be exposed")
	}
	if rec, body := serve(t, nil, "/internal/v1/users/8"); rec.Code != 404 || body["code"] != "USER_NOT_FOUND" {
		t.Fatalf("missing: %d %v", rec.Code, body)
	}
	if rec, _ := serve(t, nil, "/internal/v1/users/abc"); rec.Code != 400 {
		t.Fatalf("bad id: %d", rec.Code)
	}
}

func TestGetUserRequiresToken(t *testing.T) {
	tokens := []string{"s1", "s2"}
	if rec, _ := serve(t, tokens, "/internal/v1/users/7"); rec.Code != 401 {
		t.Fatalf("no token: %d", rec.Code)
	}
	if rec, _ := serve(t, tokens, "/internal/v1/users/7", "Authorization", "Bearer s2"); rec.Code != 200 {
		t.Fatalf("bearer: %d", rec.Code)
	}
	if rec, _ := serve(t, tokens, "/internal/v1/users/7", "X-Internal-Token", "nope"); rec.Code != 401 {
		t.Fatalf("wrong token: %d", rec.Code)
	}
}

type fakeSessions struct{}

func (fakeSessions) ValidateSession(_ context.Context, req dto.ValidateRequest) (dto.ValidateResponse, error) {
	if req.SessionToken == "access-7" {
		return dto.ValidateResponse{Valid: true, UserId: "7"}, nil
	}
	return dto.ValidateResponse{Valid: false}, nil
}

func TestGetProfile(t *testing.T) {
	rec, body := serve(t, nil, "/api/v1/users/7/profile", "Authorization", "Bearer access-7")
	if rec.Code != 200 || body["username"] != "quanluu" || body["name"] != "Quan Luu" || body["active"] != true {
		t.Fatalf("%d %v", rec.Code, body)
	}
	for _, k := range []string{"email", "phone", "address", "password"} {
		if _, ok := body[k]; ok {
			t.Fatalf("profile must not expose %s", k)
		}
	}
	if rec, _ := serve(t, nil, "/api/v1/users/7/profile"); rec.Code != 401 {
		t.Fatalf("profile without token must be rejected even when internal tokens are unset: %d", rec.Code)
	}
	if rec, _ := serve(t, nil, "/api/v1/users/7/profile", "Authorization", "Bearer expired"); rec.Code != 401 {
		t.Fatalf("invalid access token: %d", rec.Code)
	}
	if rec, _ := serve(t, nil, "/api/v1/users/7/profile", "X-Internal-Token", "anything"); rec.Code != 401 {
		t.Fatalf("internal token must not be accepted when none are configured: %d", rec.Code)
	}
	if rec, _ := serve(t, []string{"s1"}, "/api/v1/users/7/profile", "X-Internal-Token", "s1"); rec.Code != 200 {
		t.Fatalf("configured internal token: %d", rec.Code)
	}
	if rec, _ := serve(t, []string{"s1"}, "/api/v1/users/8/profile", "Authorization", "Bearer access-7"); rec.Code != 404 {
		t.Fatalf("missing user: %d", rec.Code)
	}
}
