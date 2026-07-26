package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/user-service/utils"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/copy"
	"github.com/JIeeiroSst/utils/geared_id"
)

type Service struct {
	userRepo output.UserRepository
	hasher   output.Hasher
	cache    expire.CacheHelper
}

func New(userRepo output.UserRepository, hasher output.Hasher, cache expire.CacheHelper) *Service {
	return &Service{userRepo: userRepo, hasher: hasher, cache: cache}
}

func (s *Service) SignUp(ctx context.Context, req dto.SignUpRequest) (dto.SignUpResponse, error) {
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
		return dto.SignUpResponse{}, domain.ErrUserAlready
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

	var resp dto.SignUpResponse
	if err := copy.CopyObject(&created, &resp.User); err != nil {
		return dto.SignUpResponse{}, err
	}
	resp.Message = "success"
	return resp, nil
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

	var resp dto.UpdateProfileResponse
	if err := copy.CopyObject(&updated, &resp.User); err != nil {
		return dto.UpdateProfileResponse{Message: "failed"}, err
	}
	resp.Message = "success"
	return resp, nil
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

	var resp dto.FindUserResponse
	if err := copy.CopyObject(&user, &resp.User); err != nil {
		return dto.FindUserResponse{}, err
	}
	return resp, nil
}
