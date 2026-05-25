package candidate

import (
	"context"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
)

//go:generate mockgen -source=repository.go -destination=../../mock/candidate_repo_mock.go
type Repository interface {
	Save(ctx context.Context, c *Candidate) error
	Update(ctx context.Context, c *Candidate) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Candidate, error)
	FindByEmail(ctx context.Context, email string) (*Candidate, error)
	FindAll(ctx context.Context, filter Filter) (shared.PaginatedResult[*Candidate], error)
	FindSimilar(ctx context.Context, embedding []float32, limit int) ([]*Candidate, error) // vector search
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status Status) error
}
