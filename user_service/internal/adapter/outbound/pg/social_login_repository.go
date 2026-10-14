package pg

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"gorm.io/gorm"
)

type SocialLoginRepository struct {
	db *gorm.DB
}

func NewSocialLoginRepository(db *gorm.DB) *SocialLoginRepository {
	return &SocialLoginRepository{db: db}
}

func (r *SocialLoginRepository) Find(ctx context.Context, provider, subject string) (domain.SocialLogin, error) {
	var l domain.SocialLogin
	res := r.db.WithContext(ctx).Where("provider = ? AND subject = ?", provider, subject).Limit(1).Find(&l)
	if res.Error != nil {
		return domain.SocialLogin{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.SocialLogin{}, domain.ErrUserNotExist
	}
	return l, nil
}

func (r *SocialLoginRepository) Create(ctx context.Context, login domain.SocialLogin) error {
	return r.db.WithContext(ctx).Create(&login).Error
}
