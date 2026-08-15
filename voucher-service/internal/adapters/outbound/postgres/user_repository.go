package postgres

import (
	"context"
	"time"

	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/user"
	"gorm.io/gorm"
)

type userModel struct {
	ID           string    `gorm:"column:id;primaryKey"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	Role         string    `gorm:"column:role"`
	CorporateID  *string   `gorm:"column:corporate_id"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (userModel) TableName() string { return "users" }

func userToModel(u *user.User) *userModel {
	m := &userModel{
		ID:           u.ID.String(),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		Status:       string(u.Status),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	if u.CorporateID != nil {
		s := u.CorporateID.String()
		m.CorporateID = &s
	}
	return m
}

func userFromModel(m *userModel) (*user.User, error) {
	id, err := shared.ParseUserID(m.ID)
	if err != nil {
		return nil, err
	}
	u := &user.User{
		ID:           id,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Role:         user.Role(m.Role),
		Status:       user.Status(m.Status),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.CorporateID != nil {
		cid, err := shared.ParseCorporateID(*m.CorporateID)
		if err != nil {
			return nil, err
		}
		u.CorporateID = &cid
	}
	return u, nil
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) authapp.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Create(userToModel(u)).Error
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return userFromModel(&m)
}

func (r *UserRepository) FindByID(ctx context.Context, id shared.UserID) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return userFromModel(&m)
}
