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
	s.refreshCache(ctx, updated.Id)

	return dto.UpdateProfileResponse{User: toDTOUser(updated), Message: "success"}, nil
}

func (s *Service) LockAccount(ctx context.Context, req dto.LockAccountRequest) (dto.LockAccountResponse, error) {
	if err := s.userRepo.LockAccount(ctx, int(req.Id)); err != nil {
		return dto.LockAccountResponse{Message: "failed"}, err
	}
	s.refreshCache(ctx, int(req.Id))
	return dto.LockAccountResponse{Message: "success"}, nil
}

func (s *Service) FindUser(ctx context.Context, req dto.FindUserRequest) (dto.FindUserResponse, error) {
	if req.UserId > 0 {
		user, err := s.findByID(ctx, int(req.UserId))
		if err != nil {
			return dto.FindUserResponse{}, err
		}
		return dto.FindUserResponse{Users: []*dto.User{toDTOUser(user)}, Total: 1}, nil
	}
	if req.UserId < 0 {
		return dto.FindUserResponse{}, domain.ErrInvalidRequest
	}

	filter := domain.UserFilter{}
	if req.Username != nil {
		filter.Username = *req.Username
	}
	if req.Email != nil {
		filter.Email = *req.Email
	}
	if req.Page != nil {
		filter.Page = int(*req.Page)
	}
	if req.Limit != nil {
		filter.Limit = int(*req.Limit)
	}
	users, total, err := s.userRepo.SearchUsers(ctx, filter.Normalize())
	if err != nil {
		return dto.FindUserResponse{}, err
	}
	out := dto.FindUserResponse{Users: make([]*dto.User, 0, len(users)), Total: int32(total)}
	for _, u := range users {
		out.Users = append(out.Users, toDTOUser(u))
	}
	return out, nil
}

func (s *Service) findByID(ctx context.Context, id int) (domain.User, error) {
	key := fmt.Sprintf(domain.UserCacheKey, id)
	var user domain.User
	if s.cache != nil {
		if cached, err := s.cache.GetInterface(ctx, key); err == nil {
			if raw, ok := cached.(string); ok {
				_ = json.Unmarshal([]byte(raw), &user)
			}
		}
	}
	if user.Id != 0 {
		return user, nil
	}
	fromDB, err := s.userRepo.FindUser(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	s.cacheUser(ctx, fromDB)
	return fromDB, nil
}

func (s *Service) refreshCache(ctx context.Context, id int) {
	if s.cache == nil {
		return
	}
	fromDB, err := s.userRepo.FindUser(ctx, id)
	if err != nil {
		return
	}
	s.cacheUser(ctx, fromDB)
}

func (s *Service) cacheUser(ctx context.Context, user domain.User) {
	if s.cache == nil {
		return
	}
	if payload, err := json.Marshal(user); err == nil {
		_ = s.cache.SetInterface(ctx, fmt.Sprintf(domain.UserCacheKey, user.Id), string(payload), time.Hour)
	}
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
