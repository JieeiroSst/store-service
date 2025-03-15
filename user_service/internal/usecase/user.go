package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/token"
	"github.com/JIeeiroSst/user-service/utils"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/copy"
	"github.com/JIeeiroSst/utils/geared_id"
)

type Users interface {
	Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
	SignUp(ctx context.Context, req dto.SignUpRequest) (dto.SignUpResponse, error)
	UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (dto.UpdateProfileResponse, error)
	LockAccount(ctx context.Context, req dto.LockAccountRequest) (dto.LockAccountResponse, error)
	FindUser(ctx context.Context, req dto.FindUserRequest) (dto.FindUserResponse, error)
	Authentication(ctx context.Context, req dto.AuthenticationRequest) (dto.AuthenticationResponse, error)
	ValidateSession(ctx context.Context, in dto.ValidateRequest) (dto.ValidateResponse, error)
	RefreshToken(ctx context.Context, in dto.RefreshRequest) (dto.RefreshResponse, error)
	Logout(ctx context.Context, req dto.LogoutRequest) (dto.LogoutResponse, error)
}

type UserUsecase struct {
	UserRepo repository.Users
	Hash     hash.Hash
	Token    token.Tokens
	Cache    expire.CacheHelper
	store    *SessionStore
}

func NewUsercase(UserRepo repository.Users, Hash hash.Hash,
	Token token.Tokens) *UserUsecase {
	return &UserUsecase{
		UserRepo: UserRepo,
		Hash:     Hash,
		Token:    Token,
		store:    NewSessionStore(),
	}
}

type SessionStore struct {
	sessions      map[string]string
	refreshTokens map[string]string
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions:      make(map[string]string),
		refreshTokens: make(map[string]string),
	}
}

func (u *UserUsecase) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	var user model.Users
	if err := copy.CopyObject(&req, &user); err != nil {
		return dto.LoginResponse{}, err
	}
	userId, hashPassword, err := u.UserRepo.CheckAccount(ctx, user)
	if err != nil {
		return dto.LoginResponse{}, errors.New("user does not exist")
	}
	if checkPass := u.Hash.CheckPassowrd(user.Password, hashPassword); checkPass != nil {
		return dto.LoginResponse{}, errors.New("password entered incorrectly")
	}
	sessionToken, _ := u.Token.GenerateToken(user.Username)
	refreshToken, _ := u.Token.GenerateToken(fmt.Sprintf("%s-refresh", user.Username))
	u.store.sessions[sessionToken] = strconv.Itoa(userId)
	u.store.refreshTokens[refreshToken] = strconv.Itoa(userId)

	return dto.LoginResponse{
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiryTime:   time.Now().Add(time.Hour * 24).Unix(),
	}, nil
}

func (u *UserUsecase) Logout(ctx context.Context, req dto.LogoutRequest) (dto.LogoutResponse, error) {
	_, exists := u.store.sessions[req.SessionToken]
	if !exists {
		return dto.LogoutResponse{Message: "false"}, nil
	}

	delete(u.store.sessions, req.SessionToken)
	return dto.LogoutResponse{Message: "true"}, nil
}

func (u *UserUsecase) ValidateSession(ctx context.Context, req dto.ValidateRequest) (dto.ValidateResponse, error) {
	userId, exists := u.store.sessions[req.SessionToken]
	if !exists {
		return dto.ValidateResponse{Valid: false}, nil
	}

	return dto.ValidateResponse{
		Valid:  true,
		UserId: userId,
	}, nil
}

func (u *UserUsecase) RefreshToken(ctx context.Context, req dto.RefreshRequest) (dto.RefreshResponse, error) {
	userId, exists := u.store.refreshTokens[req.RefreshToken]
	if !exists {
		return dto.RefreshResponse{}, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	newSessionToken, _ := u.Token.GenerateToken(req.Username)
	newRefreshToken, _ := u.Token.GenerateToken(fmt.Sprintf("%s-refresh", req.Username))

	u.store.sessions[newSessionToken] = userId
	u.store.refreshTokens[newRefreshToken] = userId
	delete(u.store.refreshTokens, req.RefreshToken)

	return dto.RefreshResponse{
		NewSessionToken: newSessionToken,
		NewRefreshToken: newRefreshToken,
		ExpiryTime:      time.Now().Add(time.Hour).Unix(),
	}, nil
}

func (u *UserUsecase) SignUp(ctx context.Context, req dto.SignUpRequest) (dto.SignUpResponse, error) {
	if err := utils.CheckEmail(req.Email); err != nil {
		return dto.SignUpResponse{}, err
	}
	if err := utils.CheckPassword(req.Password); err != nil {
		return dto.SignUpResponse{}, err
	}

	var user model.Users
	if err := copy.CopyObject(&req, &user); err != nil {
		return dto.SignUpResponse{}, err
	}

	check := u.UserRepo.CheckAccountExists(ctx, user)
	if check != nil {
		return dto.SignUpResponse{}, common.UserAlready
	}
	hashPassword, err := u.Hash.HashPassword(user.Password)
	if err != nil {
		return dto.SignUpResponse{}, common.HashPasswordFailed
	}
	account := model.Users{
		Id:         geared_id.GearedIntID(),
		Username:   user.Username,
		Password:   hashPassword,
		Email:      user.Email,
		Name:       user.Name,
		Sex:        user.Sex,
		Phone:      user.Phone,
		Checked:    true,
		CreateTime: time.Now(),
	}
	resp, err := u.UserRepo.CreateAccount(ctx, account)
	if err != nil {
		return dto.SignUpResponse{}, err
	}

	var respDto dto.SignUpResponse
	if err := copy.CopyObject(&resp, &respDto.User); err != nil {
		return dto.SignUpResponse{}, err
	}
	respDto.Message = "success"

	return respDto, nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (dto.UpdateProfileResponse, error) {
	var user model.Users
	if err := copy.CopyObject(&req, &user); err != nil {
		return dto.UpdateProfileResponse{Message: "failed"}, err
	}

	user, err := u.UserRepo.UpdateProfile(ctx, user)
	if err != nil {
		return dto.UpdateProfileResponse{Message: "failed"}, err
	}

	var resp dto.UpdateProfileResponse
	if err := copy.CopyObject(&user, &resp.User); err != nil {
		return dto.UpdateProfileResponse{Message: "failed"}, err
	}
	resp.Message = "success"
	return resp, nil
}

func (u *UserUsecase) LockAccount(ctx context.Context, req dto.LockAccountRequest) (dto.LockAccountResponse, error) {
	if err := u.UserRepo.LockAccount(ctx, int(req.Id)); err != nil {
		return dto.LockAccountResponse{Message: "failed"}, err
	}
	return dto.LockAccountResponse{Message: "success"}, nil
}

func (u *UserUsecase) FindUser(ctx context.Context, req dto.FindUserRequest) (dto.FindUserResponse, error) {
	var (
		users *model.Users
	)
	key := fmt.Sprintf(common.UserKey, req.UserId)

	userInterface, err := u.Cache.GetInterface(ctx, key)
	if err == redis.Nil {
		usersDB, errDB := u.UserRepo.FindUser(ctx, int(req.UserId))
		if errDB != nil {
			return dto.FindUserResponse{}, err
		}
		u.Cache.SetInterface(ctx, key, usersDB, time.Hour)
	} else {
		users = userInterface.(*model.Users)
	}

	var userDto dto.FindUserResponse
	if err := copy.CopyObject(&users, &userDto.User); err != nil {
		return dto.FindUserResponse{}, err
	}

	return userDto, nil
}

func (d *UserUsecase) Authentication(ctx context.Context, req dto.AuthenticationRequest) (dto.AuthenticationResponse, error) {
	strArr := strings.Split(req.Token, " ")
	parseToken, err := d.Token.ParseToken(strArr[1])
	if err != nil {
		return dto.AuthenticationResponse{}, err
	}
	if strings.Compare(parseToken.Username, req.Username) != 0 {
		return dto.AuthenticationResponse{}, common.FailedTokenUsername
	}
	if _, exists := d.store.sessions[req.Token]; !exists {
		return dto.AuthenticationResponse{}, common.FailedToken
	}

	var resp dto.AuthenticationResponse
	resp.Message = "success"
	return resp, nil
}
