package persistence

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/JIeeiroSst/point-service/helpers"
	"gorm.io/gorm"
)

type RewardDiscountRepositoryImpl struct {
	db *gorm.DB
}

func (r *RewardDiscountRepositoryImpl) Create(ctx context.Context, data entity.RewardDiscount) error {
	if err := r.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}

func (r *RewardDiscountRepositoryImpl) Update(ctx context.Context, data entity.RewardDiscount) error {
	if err := r.db.Model(entity.ConvertedRewardPoint{}).Where("reward_discount_id", data.RewardDiscountID).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *RewardDiscountRepositoryImpl) GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*entity.ResponseEntity, error) {
	convertedRewardPoints := []entity.RewardDiscount{}
	limit, err := strconv.ParseInt(perPage, 10, 64)
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if err != nil {
		return nil, err
	}
	isFirstPage := cursor == ""
	pointsNext := false

	query := r.db

	if cursor != "" {
		decodedCursor, err := helpers.DecodeCursor(cursor)
		if err != nil {
			return nil, err
		}
		pointsNext = decodedCursor["points_next"] == true

		operator, order := r.getPaginationOperator(pointsNext, sortOrder)
		whereStr := fmt.Sprintf("(created_at %s ? OR (created_at = ? AND id %s ?))", operator, operator)
		query = query.Where(whereStr, decodedCursor["created_at"], decodedCursor["created_at"], decodedCursor["id"])
		if order != "" {
			sortOrder = order
		}
	}

	query.Order("created_at " + sortOrder).Limit(int(limit) + 1).Find(&convertedRewardPoints)
	hasPagination := len(convertedRewardPoints) > int(limit)

	if hasPagination {
		convertedRewardPoints = convertedRewardPoints[:limit]
	}

	if !isFirstPage && !pointsNext {
		convertedRewardPoints = r.reverse(convertedRewardPoints)
	}

	pageInfo := r.calculatePagination(isFirstPage, hasPagination, int(limit), convertedRewardPoints, pointsNext)

	response := entity.ResponseEntity{
		Success:    true,
		Data:       convertedRewardPoints,
		Pagination: pageInfo,
	}

	return &response, nil
}

func (r *RewardDiscountRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.RewardDiscount, error) {
	var resp entity.RewardDiscount

	tx := r.db.Model(entity.RewardDiscount{RewardDiscountID: id}).First(&resp)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &resp, nil
}

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

func (r *RewardDiscountRepositoryImpl) reverse(s []entity.RewardDiscount) []entity.RewardDiscount {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
