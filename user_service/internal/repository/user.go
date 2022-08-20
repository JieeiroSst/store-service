package repository

import (
	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

type Users interface {
	CheckAccount(user model.Users) (int, string, error)
	CheckAccountExists(user model.Users) error
	CreateAccount(user model.Users) error
	FindUser(userId int) (model.Users, error)
	LockAccount(id int) error
	UpdateProfile(id int, user model.Users) error
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (d *UserRepository) UpdateProfile(id int, user model.Users) error {
	err := d.db.Model(model.Users{}).Where("id = ? ", id).Updates(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *UserRepository) LockAccount(id int) error {
	err := d.db.Model(&model.Users{}).Where("id = ?", id).Update("checked", false).Error
	if err != nil {
		return common.LockAccountFailed
	}
	return nil
}

func (d *UserRepository) FindUser(userId int) (model.Users, error) {
	var user model.Users
	err := d.db.Preload("Roles").Where("id = ?", userId).Find(&user).Error
	if err != nil {
		return model.Users{}, common.NotFound
	}
	return user, nil
}

func (d *UserRepository) CheckAccount(user model.Users) (int, string, error) {
	var result model.Users
	r := d.db.Where("username = ?", user.Username).Limit(1).Find(&result)

	if r.Error != nil {
		return -1, "", r.Error
	}

	if result.Id == 0 {
		return -1, "", common.UserNotExist
	}
	return result.Id, result.Password, nil
}

func (d *UserRepository) CheckAccountExists(user model.Users) error {
	var result model.Users
	r := d.db.Where("username = ?", user.Username).Limit(1).Find(&result)
	if r.Error != nil {
		return r.Error
	}

	if result.Id != 0 {
		return common.UserExist
	}
	return nil
}

func (d *UserRepository) CreateAccount(user model.Users) error {
	if err := d.db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}
