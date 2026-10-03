package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/user-service/utils"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/copy"
	"github.com/JIeeiroSst/utils/geared_id"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	userRepo   output.UserRepository
	hasher     output.Hasher
	cache      expire.CacheHelper
	authorizer output.Authorizer
}

func New(userRepo output.UserRepository, hasher output.Hasher, cache expire.CacheHelper, authorizer output.Authorizer) *Service {
	return &Service{userRepo: userRepo, hasher: hasher, cache: cache, authorizer: authorizer}
}

func (s *Service) SignUp(ctx context.Context, req dto.SignUpRequest) (dto.SignUpResponse, error) {
	if req.Username == "" {
		return dto.SignUpResponse{}, domain.ErrUsernameRequired
	}
	if err := utils.CheckEmail(req.Email); err != nil {
		return dto.SignUpResponse{}, err
	}
	if err := utils.CheckPassword(req.Password); err != nil {
		return dto.SignUpResponse{}, err
	}

	var user domain.User
	if err := copy.CopyObject(&req, &user); err != nil {
		return dto.SignUpResponse{}, err
	}

	if err := s.userRepo.CheckAccountExists(ctx, user); err != nil {
		return dto.SignUpResponse{}, err
	}
	hashedPassword, err := s.hasher.HashPassword(user.Password)
	if err != nil {
		return dto.SignUpResponse{}, domain.ErrHashPasswordFailed
	}

	account := domain.User{
		Id:         geared_id.GearedIntID(),
		Username:   user.Username,
		Password:   hashedPassword,
		Email:      user.Email,
		Name:       user.Name,
		Sex:        user.Sex,
		Phone:      user.Phone,
		Checked:    true,
		CreateTime: time.Now(),
	}
	created, err := s.userRepo.CreateAccount(ctx, account)
	if err != nil {
		return dto.SignUpResponse{}, err
	}

	if _, err := s.authorizer.AssignRoles(ctx, created.Id, domain.DefaultRole); err != nil {
		if delErr := s.userRepo.DeleteAccount(ctx, created.Id); delErr != nil {
			log.Printf("SignUp: rollback account %d failed: %v", created.Id, delErr)
		}
		return dto.SignUpResponse{}, fmt.Errorf("%w: %v", domain.ErrAssignRoleFailed, err)
	}

	return dto.SignUpResponse{User: toDTOUser(created), Message: "success"}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (dto.UpdateProfileResponse, error) {
	var user domain.User
	if err := copy.CopyObject(&req, &user); err != nil {
		return dto.UpdateProfileResponse{Message: "failed"}, err
	}

	updated, err := s.userRepo.UpdateProfile(ctx, user)
	if err != nil {
		return dto.UpdateProfileResponse{Message: "failed"}, err
	}

	return dto.UpdateProfileResponse{User: toDTOUser(updated), Message: "success"}, nil
}

func (s *Service) LockAccount(ctx context.Context, req dto.LockAccountRequest) (dto.LockAccountResponse, error) {
	if err := s.userRepo.LockAccount(ctx, int(req.Id)); err != nil {
		return dto.LockAccountResponse{Message: "failed"}, err
	}
	return dto.LockAccountResponse{Message: "success"}, nil
}

func (s *Service) FindUser(ctx context.Context, req dto.FindUserRequest) (dto.FindUserResponse, error) {
	key := fmt.Sprintf(domain.UserCacheKey, req.UserId)

	var user domain.User
	if cached, err := s.cache.GetInterface(ctx, key); err == nil {
		if raw, ok := cached.(string); ok {
			_ = json.Unmarshal([]byte(raw), &user)
		}
	}

	if user.Id == 0 {
		fromDB, err := s.userRepo.FindUser(ctx, int(req.UserId))
		if err != nil {
			return dto.FindUserResponse{}, err
		}
		user = fromDB
		if payload, err := json.Marshal(user); err == nil {
			_ = s.cache.SetInterface(ctx, key, string(payload), time.Hour)
		}
	}

	return dto.FindUserResponse{User: toDTOUser(user)}, nil
}

func toDTOUser(u domain.User) *dto.User {
	out := &dto.User{
		Id:         int32(u.Id),
		Username:   u.Username,
		Email:      u.Email,
		Name:       u.Name,
		Phone:      u.Phone,
		Address:    u.Address,
		Sex:        u.Sex,
		Checked:    u.Checked,
		CreateTime: timestamppb.New(u.CreateTime),
	}
	if !u.UpdateTime.IsZero() {
		out.UpdateTime = timestamppb.New(u.UpdateTime)
	}
	for _, r := range u.Roles {
		out.Roles = append(out.Roles, &dto.Role{Id: int32(r.Id), Name: r.Name})
	}
	return out
}
