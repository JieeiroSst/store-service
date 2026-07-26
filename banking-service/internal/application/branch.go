package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type branchService struct {
	repo port.BranchRepository
}

func NewBranchService(repo port.BranchRepository) port.BranchUsecase {
	return &branchService{repo: repo}
}

func (s *branchService) CreateBranch(ctx context.Context, branch *model.Branch) (*model.Branch, error) {
	if err := s.repo.Create(ctx, branch); err != nil {
		return nil, err
	}
	return branch, nil
}

func (s *branchService) GetBranch(ctx context.Context, id int) (*model.Branch, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *branchService) ListBranches(ctx context.Context) ([]model.Branch, error) {
	return s.repo.List(ctx)
}

func (s *branchService) UpdateBranch(ctx context.Context, branch *model.Branch) (*model.Branch, error) {
	if _, err := s.repo.GetByID(ctx, branch.BranchID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, branch); err != nil {
		return nil, err
	}
	return branch, nil
}

func (s *branchService) DeleteBranch(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
