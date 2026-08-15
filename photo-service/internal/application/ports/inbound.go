package ports

import (
	"context"

	"github.com/JIeeiroSst/photo-service/internal/domain"
)

type ComposeImagesCommand struct {
	Sources        []domain.ImageSource
	Layout         domain.LayoutConfig
	IdempotencyKey string
}

type ComposeImageUseCase interface {
	ComposeImages(ctx context.Context, cmd ComposeImagesCommand) (*domain.CompositionJob, error)
	GetComposition(ctx context.Context, id string) (*domain.CompositionJob, error)
}
