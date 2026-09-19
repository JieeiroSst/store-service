package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) port.TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) GetOrCreateByNames(ctx context.Context, names []string) ([]model.Tag, error) {
	tags := make([]model.Tag, 0, len(names))
	for _, name := range names {
		var tag model.Tag
		err := r.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			tag = model.Tag{ID: uuid.NewString(), Name: name}
			if err := r.db.WithContext(ctx).
				Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
				Create(&tag).Error; err != nil {
				return nil, err
			}
			if err := r.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error; err != nil {
				return nil, err
			}
		case err != nil:
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *tagRepository) AttachToPost(ctx context.Context, postID string, tagIDs []string) error {
	rows := make([]model.PostTag, len(tagIDs))
	for i, tagID := range tagIDs {
		rows[i] = model.PostTag{PostID: postID, TagID: tagID}
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&rows).Error
}

func (r *tagRepository) ListNamesByPost(ctx context.Context, postID string) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).
		Table("tags").
		Joins("JOIN post_tags ON post_tags.tag_id = tags.id").
		Where("post_tags.post_id = ?", postID).
		Pluck("tags.name", &names).Error
	return names, err
}
