package repository

import (
	"database/sql"

	"github.com/JIeeiroSst/workflow-service/model"
)

type ActiveUsers interface {
	InsertActiveUser(user model.ActiveUser) error
}

type ActiveUsersRepo struct {
	DB *sql.DB
}

func NewActiveUsersRepo(db *sql.DB) *ActiveUsersRepo {
	return &ActiveUsersRepo{
		DB: db,
	}
}

func (r *ActiveUsersRepo) InsertActiveUser(user []model.ActiveUser) error {
	return nil
}
