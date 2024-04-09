package usecase

import (
	"context"

	"github.com/JIeerioSst/subsidies-service/dto"
	"github.com/JIeerioSst/subsidies-service/internal/redis"
	"github.com/JIeerioSst/subsidies-service/internal/repository"
)

type Programs interface {
	Save(ctx context.Context, Program dto.Program) error
}
type ProgramsUsecase struct {
	programRepo  repository.ProgramRepo
	cacheProgram redis.CacheProgram
}

func NewProgramsUsecase(programRepo repository.ProgramRepo,
	cacheProgram redis.CacheProgram) Programs {
	return &ProgramsUsecase{
		programRepo:  programRepo,
		cacheProgram: cacheProgram,
	}
}

func (u *ProgramsUsecase) Save(ctx context.Context, Program dto.Program) error {
	
	return nil
}
