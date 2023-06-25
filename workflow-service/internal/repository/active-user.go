package repository

import (
	"database/sql"

	"github.com/JIeeiroSst/workflow-service/model"
)

type ActiveUsers interface {
	InsertActiveUser(user model.ActiveUser) error
	UpdateActiveUser(id string, user model.ActiveUser) error
}

type ActiveUsersRepo struct {
	DB *sql.DB
}

func NewActiveUsersRepo(db *sql.DB) *ActiveUsersRepo {
	return &ActiveUsersRepo{
		DB: db,
	}
}

func (r *ActiveUsersRepo) InsertActiveUser(user model.ActiveUser) error {
	sqlStatement := `
		INSERT INTO active_users (id, key, value,user_peding,user_approve,user_reject,status,create_at,update_at,delete_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, $9,$10)`
	_, err := r.DB.Exec(sqlStatement, user.ID, user.Key, user.Value, user.UserPeding, user.UserApprove, user.UserReject, user.Status, user.CreateAt, user.UpdateAt, user.DeleteAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *ActiveUsersRepo) UpdateActiveUser(id string, user model.ActiveUser) error {
	sqlStatement := `
				UPDATE active_users
				SET key = $2, value =$3, user_peding=$4, user_approve=$5, user_reject=$6,
					status=$7, create_at=$8, update_at=$9, delete_at=$10
				WHERE id = $1;`
	_, err := r.DB.Exec(sqlStatement, user.ID, user.Key, user.Value, user.UserPeding, user.UserApprove, user.UserReject, user.Status, user.CreateAt, user.UpdateAt, user.DeleteAt)
	if err != nil {
		return err
	}
	return nil
}
