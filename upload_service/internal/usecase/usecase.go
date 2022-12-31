package usecase

import (
	"github.com/JIeeiroSst/upload-service/internal/repository"
	"github.com/JIeeiroSst/upload-service/pkg/snowflake"
)

type Usecase struct {
	Uploads
}

type Dependency struct {
	uploadRepo repository.Uploads
	snowflake  snowflake.SnowflakeData
}

func NewUsecase(deps Dependency) *Usecase {
	uploadUsecase := NewUploadUsecase(deps.uploadRepo, deps.snowflake)

	return &Usecase{
		Uploads: uploadUsecase,
	}
}
