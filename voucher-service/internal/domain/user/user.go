package user

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleRetail          Role = "retail"
	RoleCorporateAdmin  Role = "corporate_admin"
	RoleSystemAdmin     Role = "system_admin"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type User struct {
	ID           shared.UserID
	Email        string
	PasswordHash string
	Role         Role
	CorporateID  *shared.CorporateID
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, plaintextPassword string, role Role, now time.Time) (*User, error) {
	if email == "" || plaintextPassword == "" {
		return nil, ErrInvalidUser
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:           shared.NewUserID(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Status:       StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) VerifyPassword(plaintext string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintext)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (u *User) IsActive() bool { return u.Status == StatusActive }
