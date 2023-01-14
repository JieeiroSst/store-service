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
	Repo      repository.Repositories
	Snowflake snowflake.SnowflakeData
	UploadApi uploadAPI.UploadApi
}

func NewUsecase(deps Dependency) *Usecase {
	uploadUsecase := NewUploadUsecase(deps.Repo.Uploads, deps.Snowflake, deps.UploadApi)
	return &Usecase{
		Uploads: uploadUsecase,
	}
}
