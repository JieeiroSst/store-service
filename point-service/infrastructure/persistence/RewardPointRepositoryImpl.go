package persistence

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/JIeeiroSst/point-service/helpers"
	"gorm.io/gorm"
)

type RewardPointRepositoryImpl struct {
	db *gorm.DB
}

func (r *RewardPointRepositoryImpl) Create(ctx context.Context, data entity.RewardDiscount) error
func (r *RewardPointRepositoryImpl) Update(ctx context.Context, data entity.RewardDiscount) error
func (r *RewardPointRepositoryImpl) GetAll(ctx context.Context, perPage int, sortOrder, cursor string) ([]entity.ResponseEntity, error)
func (r *RewardPointRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.RewardDiscount, error)

func (r *RewardPointRepositoryImpl) getPaginationOperator(pointsNext bool, sortOrder string) (string, string) {
	if pointsNext && sortOrder == "asc" {
		return ">", ""
	}
	if pointsNext && sortOrder == "desc" {
		return "<", ""
	}
	if !pointsNext && sortOrder == "asc" {
		return "<", "desc"
	}
	if !pointsNext && sortOrder == "desc" {
		return ">", "asc"
	}

	return "", ""
}

func (r *RewardPointRepositoryImpl) calculatePagination(isFirstPage bool, hasPagination bool, limit int, convertedRewardPoints []entity.RewardPoint, pointsNext bool) helpers.PaginationInfo {
	pagination := helpers.PaginationInfo{}
	nextCur := helpers.Cursor{}
	prevCur := helpers.Cursor{}
	if isFirstPage {
		if hasPagination {
			nextCur := helpers.CreateCursor(convertedRewardPoints[limit-1].RewardPointsId, convertedRewardPoints[limit-1].CreatedAt, true)
			pagination = helpers.GeneratePager(nextCur, nil)
		}
	} else {
		if pointsNext {
			// if pointing next, it always has prev but it might not have next
			if hasPagination {
				nextCur = helpers.CreateCursor(convertedRewardPoints[limit-1].RewardPointsId, convertedRewardPoints[limit-1].CreatedAt, true)
			}
			prevCur = helpers.CreateCursor(convertedRewardPoints[0].RewardPointsId, convertedRewardPoints[0].CreatedAt, false)
			pagination = helpers.GeneratePager(nextCur, prevCur)
		} else {
			// this is case of prev, there will always be nest, but prev needs to be calculated
			nextCur = helpers.CreateCursor(convertedRewardPoints[limit-1].RewardPointsId, convertedRewardPoints[limit-1].CreatedAt, true)
			if hasPagination {
				prevCur = helpers.CreateCursor(convertedRewardPoints[0].RewardPointsId, convertedRewardPoints[0].CreatedAt, false)
			}
			pagination = helpers.GeneratePager(nextCur, prevCur)
		}
	}
	return pagination
}

func (r *RewardPointRepositoryImpl) Reverse(s []entity.RewardPoint) []entity.RewardPoint {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
