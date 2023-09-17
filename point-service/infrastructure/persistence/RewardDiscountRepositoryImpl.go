package persistence

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/JIeeiroSst/point-service/helpers"
	"gorm.io/gorm"
)

type RewardDiscountRepositoryImpl struct {
	db *gorm.DB
}

func (r *RewardDiscountRepositoryImpl) Create(ctx context.Context, data entity.RewardPoint) error
func (r *RewardDiscountRepositoryImpl) Update(ctx context.Context, data entity.RewardPoint) error
func (r *RewardDiscountRepositoryImpl) GetAll(ctx context.Context, perPage int, sortOrder, cursor string) (entity.ResponseEntity, error)
func (r *RewardDiscountRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.RewardPoint, error)

func (r *RewardDiscountRepositoryImpl) getPaginationOperator(pointsNext bool, sortOrder string) (string, string) {
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

func (r *RewardDiscountRepositoryImpl) calculatePagination(isFirstPage bool, hasPagination bool, limit int, convertedRewardPoints []entity.RewardDiscount, pointsNext bool) helpers.PaginationInfo {
	pagination := helpers.PaginationInfo{}
	nextCur := helpers.Cursor{}
	prevCur := helpers.Cursor{}
	if isFirstPage {
		if hasPagination {
			nextCur := helpers.CreateCursor(convertedRewardPoints[limit-1].RewardDiscountID, convertedRewardPoints[limit-1].CreatedAt, true)
			pagination = helpers.GeneratePager(nextCur, nil)
		}
	} else {
		if pointsNext {
			// if pointing next, it always has prev but it might not have next
			if hasPagination {
				nextCur = helpers.CreateCursor(convertedRewardPoints[limit-1].RewardDiscountID, convertedRewardPoints[limit-1].CreatedAt, true)
			}
			prevCur = helpers.CreateCursor(convertedRewardPoints[0].RewardDiscountID, convertedRewardPoints[0].CreatedAt, false)
			pagination = helpers.GeneratePager(nextCur, prevCur)
		} else {
			// this is case of prev, there will always be nest, but prev needs to be calculated
			nextCur = helpers.CreateCursor(convertedRewardPoints[limit-1].RewardDiscountID, convertedRewardPoints[limit-1].CreatedAt, true)
			if hasPagination {
				prevCur = helpers.CreateCursor(convertedRewardPoints[0].RewardDiscountID, convertedRewardPoints[0].CreatedAt, false)
			}
			pagination = helpers.GeneratePager(nextCur, prevCur)
		}
	}
	return pagination
}

func (r *RewardDiscountRepositoryImpl) Reverse(s []entity.RewardDiscount) []entity.RewardDiscount {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
