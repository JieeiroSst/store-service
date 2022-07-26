package usecase

import (
	"github.com/JIeeiroSst/post-service/internal/repository"
	"github.com/JIeeiroSst/post-service/pkg/minio"
	"github.com/JIeeiroSst/post-service/pkg/snowflake"
)

type Usecase struct {
	Categories
	News
}

type Dependency struct {
	Repos     *repository.Repositories
	Snowflake snowflake.SnowflakeData
	Minio     minio.ClientS3
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Categories: NewCategoryUsecase(deps.Repos.Categories, deps.Snowflake),
		News:NewNewsUsecase(deps.Repos.News,deps.Snowflake,deps.Repos.Medias,deps.Minio),
	}
}
