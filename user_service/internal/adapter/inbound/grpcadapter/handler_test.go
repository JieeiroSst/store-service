package grpcadapter

import (
	"context"
	"testing"
	"time"

	userServiceGrpc "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeUserService struct {
	input.UserService
	got dto.FindUserRequest
}

func (f *fakeUserService) FindUser(_ context.Context, req dto.FindUserRequest) (dto.FindUserResponse, error) {
	f.got = req
	return dto.FindUserResponse{Total: 1, Users: []*dto.User{{
		Id: 3, Username: *req.Username, CreateTime: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	}}}, nil
}

func TestFindUserMapsRequestAndResponse(t *testing.T) {
	users := &fakeUserService{}
	h := NewHandler(nil, users, nil, nil)
	name := "alice"
	limit := int32(5)
	res, err := h.FindUser(context.Background(), &userServiceGrpc.FindUserRequest{Username: &name, Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
	if users.got.Username == nil || *users.got.Username != "alice" || users.got.Limit == nil || *users.got.Limit != 5 {
		t.Fatalf("request not mapped: %+v", users.got)
	}
	if res.Total != 1 || len(res.Users) != 1 || res.Users[0].Id != 3 || res.Users[0].Username != "alice" ||
		res.Users[0].CreateTime.AsTime().Year() != 2026 {
		t.Fatalf("response not mapped: %+v", res)
	}
}

type fakeAuthService struct {
	input.AuthService
	got dto.ValidateRequest
}

func (f *fakeAuthService) ValidateSession(_ context.Context, req dto.ValidateRequest) (dto.ValidateResponse, error) {
	f.got = req
	return dto.ValidateResponse{Valid: true, UserId: "42"}, nil
}

func TestValidateSessionMapsRequestAndResponse(t *testing.T) {
	auth := &fakeAuthService{}
	h := NewHandler(auth, nil, nil, nil)
	res, err := h.ValidateSession(context.Background(), &userServiceGrpc.ValidateRequest{SessionToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if auth.got.SessionToken != "tok" || !res.Valid || res.UserId != "42" {
		t.Fatalf("got req %+v res %+v", auth.got, res)
	}
}
