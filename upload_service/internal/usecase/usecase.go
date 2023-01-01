package usecase

import (
	"github.com/JIeeiroSst/upload-service/internal/repository"
	uploadAPI "github.com/JIeeiroSst/upload-service/pkg/api"
	"github.com/JIeeiroSst/upload-service/pkg/snowflake"
)

type Usecase struct {
	Uploads
}

type Dependency struct {
	uploadRepo repository.Uploads
	snowflake  snowflake.SnowflakeData
	uploadApi  uploadAPI.UploadApi
}

func NewUsecase(deps Dependency) *Usecase {
	uploadUsecase := NewUploadUsecase(deps.uploadRepo, deps.snowflake, deps.uploadApi)
	return &Usecase{
		Uploads: uploadUsecase,
	}
}
