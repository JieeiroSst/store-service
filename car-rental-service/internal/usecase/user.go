package usecase

import (
	"context"
	"net/mail"
	"strings"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserInput struct {
	Email, Password, FirstName, LastName string
	PhoneNumber, Address, DrivingLicense string
	UserType                             model.UserType
	AllowPrivileged                      bool
}

func (u *Usecase) RegisterUser(ctx context.Context, in RegisterUserInput) (*model.User, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return nil, invalid("email is not valid")
	}
	if len(in.Password) < 8 {
		return nil, invalid("password must be at least 8 characters")
	}
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return nil, invalid("first_name and last_name are required")
	}
	userType := model.UserTypeCustomer
	if in.UserType != "" && in.UserType != model.UserTypeCustomer {
		if !in.AllowPrivileged {
			return nil, model.ErrPermissionDenied
		}
		userType = in.UserType
	}

	exists, err := u.repos.Users.EmailExists(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, model.ErrAlreadyExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		ID:             uuid.New(),
		Email:          in.Email,
		PasswordHash:   string(hash),
		FirstName:      strings.TrimSpace(in.FirstName),
		LastName:       strings.TrimSpace(in.LastName),
		PhoneNumber:    in.PhoneNumber,
		Address:        in.Address,
		DrivingLicense: in.DrivingLicense,
		UserType:       userType,
	}
	if err := u.repos.Users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *Usecase) GetUser(ctx context.Context, id string) (*model.User, error) {
	uid, err := parseID("user_id", id)
	if err != nil {
		return nil, err
	}
	return u.repos.Users.GetByID(ctx, uid)
}

type UpdateUserInput struct {
	UserID                               string
	Email, FirstName, LastName           string
	PhoneNumber, Address, DrivingLicense string
}

// UpdateUser applies the non-empty fields of the input.
func (u *Usecase) UpdateUser(ctx context.Context, in UpdateUserInput) (*model.User, error) {
	user, err := u.GetUser(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if email := strings.ToLower(strings.TrimSpace(in.Email)); email != "" && email != user.Email {
		if _, err := mail.ParseAddress(email); err != nil {
			return nil, invalid("email is not valid")
		}
		exists, err := u.repos.Users.EmailExists(ctx, email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, model.ErrAlreadyExists
		}
		user.Email = email
	}
	set := func(dst *string, v string) {
		if v = strings.TrimSpace(v); v != "" {
			*dst = v
		}
	}
	set(&user.FirstName, in.FirstName)
	set(&user.LastName, in.LastName)
	set(&user.PhoneNumber, in.PhoneNumber)
	set(&user.Address, in.Address)
	set(&user.DrivingLicense, in.DrivingLicense)
	if err := u.repos.Users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *Usecase) requireStaff(ctx context.Context, id string) (*model.User, error) {
	user, err := u.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.UserType != model.UserTypeStaff && user.UserType != model.UserTypeAdmin {
		return nil, model.ErrPermissionDenied
	}
	return user, nil
}
