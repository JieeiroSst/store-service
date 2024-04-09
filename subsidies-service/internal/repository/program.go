package repository

import (
	"context"

	"github.com/JIeerioSst/subsidies-service/model"
	"google.golang.org/appengine/log"
	"gorm.io/gorm"
)

type Programs interface {
	Save(ctx context.Context, program model.Program) error
}

type ProgramRepo struct {
	db *gorm.DB
}

func NewProgramRepo(db *gorm.DB) Programs {
	return &ProgramRepo{db: db}
}

func (r *ProgramRepo) Save(ctx context.Context, program model.Program) error {
	if err := r.db.Save(program).Error; err != nil {
		log.Errorf(ctx, "Save Program in err = %v", err.Error())
		return err
	}
	return nil
}
