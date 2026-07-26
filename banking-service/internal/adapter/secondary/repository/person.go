package repository

import (
	"context"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"gorm.io/gorm"
)

type personRepository struct {
	db *gorm.DB
}

func NewPersonRepository(db *gorm.DB) port.PersonRepository {
	return &personRepository{db: db}
}

func (r *personRepository) Create(ctx context.Context, person *model.Person) error {
	if err := r.db.WithContext(ctx).Create(person).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *personRepository) GetByID(ctx context.Context, id int) (*model.Person, error) {
	var person model.Person
	err := r.db.WithContext(ctx).First(&person, "person_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &person, nil
}

func (r *personRepository) List(ctx context.Context) ([]model.Person, error) {
	var persons []model.Person
	if err := r.db.WithContext(ctx).Find(&persons).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return persons, nil
}

func (r *personRepository) Update(ctx context.Context, person *model.Person) error {
	if err := r.db.WithContext(ctx).Model(&model.Person{}).Where("person_id = ?", person.PersonID).Updates(person).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *personRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Person{}, "person_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
