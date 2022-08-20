package usecase

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/token"

	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
)

type Users interface {
	Login(user model.Users) (int, string, error)
	SignUp(user model.Users) error
	UpdateProfile(id int, user model.Users) error
	LockAccount(id int) error
	FindUser(userId int) (*model.Users, error)

	CheckPassword(password string) error
	CheckEmail(email string) error
	CheckIP(ip string) error

	Authentication(token string, username string) error
}

type UserUsecase struct {
	UserRepo  repository.Users
	Snowflake snowflake.SnowflakeData
	Hash      hash.Hash
	Token     token.Tokens
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

func (u *UserUsecase) Login(user model.Users) (int, string, error) {
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

func (u *UserUsecase) SignUp(user model.Users) error {
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

func (u *UserUsecase) UpdateProfile(id int, user model.Users) error {
	if err := u.UserRepo.UpdateProfile(id, user); err != nil {
		return err
	}
	return nil
}

func (u *UserUsecase) LockAccount(id int) error {
	if err := u.UserRepo.LockAccount(id); err != nil {
		return err
	}
	return nil
}

func (u *UserUsecase) FindUser(userId int) (*model.Users, error) {
	users, err := u.UserRepo.FindUser(userId)
	if err != nil {
		return nil, err
	}
	return &users, nil
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

func (d *UserUsecase) Authentication(token string, username string) error {
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
