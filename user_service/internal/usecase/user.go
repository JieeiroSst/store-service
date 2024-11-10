package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/token"
	"github.com/go-redis/redis/v8"

	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Users interface {
	Login(ctx context.Context, user model.Users) (int, string, error)
	SignUp(ctx context.Context, user model.Users) error
	UpdateProfile(ctx context.Context, id int, user model.Users) error
	LockAccount(ctx context.Context, id int) error
	FindUser(ctx context.Context, userId int) (*model.Users, error)

	CheckPassword(password string) error
	CheckEmail(email string) error
	CheckIP(ip string) error

	Authentication(ctx context.Context, token string, username string) error
}

type UserUsecase struct {
	UserRepo  repository.Users
	Snowflake snowflake.SnowflakeData
	Hash      hash.Hash
	Token     token.Tokens
	Cache     expire.CacheHelper
}

func NewUsercase(UserRepo repository.Users,
	Snowflake snowflake.SnowflakeData, Hash hash.Hash,
	Token token.Tokens) *UserUsecase {
	return &UserUsecase{
		UserRepo:  UserRepo,
		Snowflake: Snowflake,
		Hash:      Hash,
		Token:     Token,
	}
}

func (u *UserUsecase) Login(ctx context.Context, user model.Users) (int, string, error) {
	id, hashPassword, err := u.UserRepo.CheckAccount(user)
	if err != nil {
		return 0, "", errors.New("user does not exist")
	}
	if checkPass := u.Hash.CheckPassowrd(user.Password, hashPassword); checkPass != nil {
		return 0, "", errors.New("password entered incorrectly")
	}
	token, _ := u.Token.GenerateToken(user.Username)
	return id, token, nil
}

func (u *UserUsecase) SignUp(ctx context.Context, user model.Users) error {
	if err := u.CheckEmail(user.Email); err != nil {
		return err
	}
	if err := u.CheckPassword(user.Password); err != nil {
		return err
	}
	check := u.UserRepo.CheckAccountExists(user)
	if check != nil {
		return common.UserAlready
	}
	hashPassword, err := u.Hash.HashPassword(user.Password)
	if err != nil {
		return common.HashPasswordFailed
	}
	account := model.Users{
		Id:         u.Snowflake.GearedID(),
		Username:   user.Username,
		Password:   hashPassword,
		Email:      user.Email,
		Name:       user.Name,
		Sex:        user.Sex,
		Phone:      user.Phone,
		Checked:    true,
		CreateTime: time.Now(),
	}
	err = u.UserRepo.CreateAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, id int, user model.Users) error {
	if err := u.UserRepo.UpdateProfile(id, user); err != nil {
		return err
	}
	return nil
}

func (u *UserUsecase) LockAccount(ctx context.Context, id int) error {
	if err := u.UserRepo.LockAccount(id); err != nil {
		return err
	}
	return nil
}

func (u *UserUsecase) FindUser(ctx context.Context, userId int) (*model.Users, error) {
	var (
		users *model.Users
	)
	key := fmt.Sprintf(common.UserKey, userId)

	userInterface, err := u.Cache.GetInterface(ctx, key)
	if err == redis.Nil {
		usersDB, errDB := u.UserRepo.FindUser(userId)
		if errDB != nil {
			return nil, err
		}
		u.Cache.SetInterface(ctx, key, usersDB, time.Hour)
	} else {
		users = userInterface.(*model.Users)
	}

	return users, nil
}

func (d *UserUsecase) CheckPassword(password string) error {
	regex := `([A-Z])\w+`
	matched, err := regexp.MatchString(regex, password)
	if !matched {
		return common.PasswordFailed
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *UserUsecase) CheckEmail(email string) error {
	regex := `^[a-z][a-z0-9_\.]{5,32}@[a-z0-9]{2,}(\.[a-z0-9]{2,4}){1,2}$`
	matched, err := regexp.MatchString(regex, email)
	if !matched {
		return common.EmailFailed
	}
	if err != nil {
		return nil
	}
	return nil
}

func (d *UserUsecase) CheckIP(ip string) error {
	regex := `/^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/`
	matched, err := regexp.MatchString(regex, ip)
	if !matched {
		return common.IPFailed
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *UserUsecase) Authentication(ctx context.Context, token string, username string) error {
	strArr := strings.Split(token, " ")
	parseToken, err := d.Token.ParseToken(strArr[1])
	if err != nil {
		return err
	}
	if strings.Compare(parseToken.Username, username) != 0 {
		return common.FailedTokenUsername
	}
	return nil
}
