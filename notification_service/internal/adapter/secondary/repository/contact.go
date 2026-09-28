package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type contactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) port.ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) Upsert(ctx context.Context, c *model.UserContact) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if c.Email != nil {
			if err := tx.Model(&model.UserContact{}).Where("email = ? AND user_id <> ?", *c.Email, c.UserID).Update("email", nil).Error; err != nil {
				return err
			}
		}
		if c.Phone != nil {
			if err := tx.Model(&model.UserContact{}).Where("phone = ? AND user_id <> ?", *c.Phone, c.UserID).Update("phone", nil).Error; err != nil {
				return err
			}
		}
		columns := []string{"updated_at"}
		if c.Email != nil {
			columns = append(columns, "email")
		}
		if c.Phone != nil {
			columns = append(columns, "phone")
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns(columns),
		}).Create(c).Error
	})
	if err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *contactRepository) Get(ctx context.Context, userID uint) (*model.UserContact, error) {
	var c model.UserContact
	if err := r.db.WithContext(ctx).First(&c, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &c, nil
}

func (r *contactRepository) userIDBy(ctx context.Context, column, value string) (uint, error) {
	var c model.UserContact
	if err := r.db.WithContext(ctx).Select("user_id").Where(column+" = ?", value).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, common.ErrUnknownContact
		}
		return 0, common.ErrDBFailed
	}
	return c.UserID, nil
}

func (r *contactRepository) UserIDByEmail(ctx context.Context, email string) (uint, error) {
	return r.userIDBy(ctx, "email", email)
}

func (r *contactRepository) UserIDByPhone(ctx context.Context, phone string) (uint, error) {
	return r.userIDBy(ctx, "phone", phone)
}
