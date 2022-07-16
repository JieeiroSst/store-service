package repository

import (
	"errors"

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
	FindAllUser() (model.Users, error)
	LockAccount(id int) error
	UpdateProfile(id int, user model.Users) error
}

func NewUserRepository(db *gorm.DB) *UserRepository{
	return &UserRepository{
		db: db,
	}
}

func (d *UserRepository) UpdateProfile(id int, user model.Users) error {
	err := d.db.Model(model.Users{}).Where("id = ? ", id).Updates(user).Error
	if err != nil {
		return errors.New("update profile user failed")
	}
	return nil
}

func (d *UserRepository) LockAccount(id int) error {
	err := d.db.Model(&model.Users{}).Where("id = ?", id).Update("checked", false).Error
	if err != nil {
		return errors.New("lock account failed")
	}
	return nil
}

func (d *UserRepository) FindAllUser() (model.Users, error) {
	var user model.Users
	err := d.db.Select("id, username").Find(&user).Error
	if err != nil {
		return model.Users{}, errors.New("")
	}
	return user, nil
}

func (d *UserRepository) CheckAccount(user model.Users) (int, string, error) {
	var result model.Users
	r := d.db.Where("username = ?", user.Username).Limit(1).Find(&result)

	if r.Error != nil {
		return -1, "", errors.New("Query error")
	}

	if result.Id == 0 {
		return -1, "", errors.New("user does not exist")
	}
	return result.Id, result.Password, nil
}

func (d *UserRepository) CheckAccountExists(user model.Users) error {
	var result model.Users
	r := d.db.Where("username = ?", user.Username).Limit(1).Find(&result)
	if r.Error != nil {
		return errors.New("query error")
	}

	if result.Id != 0 {
		return errors.New("user does exist")
	}
	return nil
}

func (d *UserRepository) CreateAccount(user model.Users) error {
	if err := d.db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}