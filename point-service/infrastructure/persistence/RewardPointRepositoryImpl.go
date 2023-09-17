package persistence

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/JIeeiroSst/point-service/helpers"
	"gorm.io/gorm"
)

type RewardPointRepositoryImpl struct {
	db *gorm.DB
}

func NewRewardPointRepositoryImpl(db *gorm.DB) *RewardPointRepositoryImpl {
	return &RewardPointRepositoryImpl{
		db: db,
	}
}

func (r *RewardPointRepositoryImpl) Create(ctx context.Context, data entity.RewardPoint) error {
	if err := r.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}

func (r *RewardPointRepositoryImpl) Update(ctx context.Context, data entity.RewardPoint) error {
	if err := r.db.Model(entity.ConvertedRewardPoint{}).Where("reward_points_id", data.RewardPointsId).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *RewardPointRepositoryImpl) GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*entity.ResponseEntity, error) {
	rewardPoints := []entity.RewardPoint{}
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

	query.Order("created_at " + sortOrder).Limit(int(limit) + 1).Find(&rewardPoints)
	hasPagination := len(rewardPoints) > int(limit)

	if hasPagination {
		rewardPoints = rewardPoints[:limit]
	}

	if !isFirstPage && !pointsNext {
		rewardPoints = r.reverse(rewardPoints)
	}

	pageInfo := r.calculatePagination(isFirstPage, hasPagination, int(limit), rewardPoints, pointsNext)

	response := entity.ResponseEntity{
		Success:    true,
		Data:       rewardPoints,
		Pagination: pageInfo,
	}

	return &response, nil
}

func (r *RewardPointRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.RewardPoint, error) {
	var resp entity.RewardPoint

	tx := r.db.Model(entity.RewardPoint{RewardPointsId: id}).First(&resp)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &resp, nil
}

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

func (r *RewardPointRepositoryImpl) reverse(s []entity.RewardPoint) []entity.RewardPoint {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
