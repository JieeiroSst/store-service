package repository

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo struct{ db *gorm.DB }

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "user_id = ?", id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &u, nil
}

func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&n).Error
	return n > 0, err
}

func (r *UserRepo) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
