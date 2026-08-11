package pg

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (d *UserRepository) UpdateProfile(ctx context.Context, user domain.User) (domain.User, error) {
	if err := d.db.Model(domain.User{}).Where("id = ? ", user.Id).Updates(user).Error; err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (d *UserRepository) LockAccount(ctx context.Context, id int) error {
	if err := d.db.Model(&domain.User{}).Where("id = ?", id).Update("checked", false).Error; err != nil {
		return domain.ErrLockAccountFailed
	}
	return nil
}

func (d *UserRepository) FindUser(ctx context.Context, userId int) (domain.User, error) {
	var user domain.User
	if err := d.db.Preload("Roles").Where("id = ?", userId).Find(&user).Error; err != nil {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (d *UserRepository) CheckAccount(ctx context.Context, user domain.User) (int, string, string, error) {
	var result domain.User
	r := d.db.Preload("Roles").Where("username = ?", user.Username).Limit(1).Find(&result)
	if r.Error != nil {
		return -1, "", "", r.Error
	}
	if result.Id == 0 {
		return -1, "", "", domain.ErrUserNotExist
	}

	role := "user"
	if len(result.Roles) > 0 {
		role = result.Roles[0].Name
	}
	return result.Id, result.Password, role, nil
}

func (d *UserRepository) CheckAccountExists(ctx context.Context, user domain.User) error {
	var result domain.User
	r := d.db.Where("username = ?", user.Username).Limit(1).Find(&result)
	if r.Error != nil {
		return r.Error
	}
	if result.Id != 0 {
		return domain.ErrUserExist
	}
	return nil
}

func (d *UserRepository) CreateAccount(ctx context.Context, user domain.User) (domain.User, error) {
	if err := d.db.Create(&user).Error; err != nil {
		return domain.User{}, err
	}
	return user, nil
}
