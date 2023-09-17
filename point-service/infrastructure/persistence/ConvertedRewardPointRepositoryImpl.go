package persistence

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/JIeeiroSst/point-service/helpers"
	"gorm.io/gorm"
)

type ConvertedRewardPointRepositoryImpl struct {
	db *gorm.DB
}

func NewConvertedRewardPointRepositoryImpl(db *gorm.DB) *ConvertedRewardPointRepositoryImpl {
	return &ConvertedRewardPointRepositoryImpl{
		db: db,
	}
}

func (t *ConvertedRewardPointRepositoryImpl) Create(ctx context.Context, data entity.ConvertedRewardPoint) error {
	if err := t.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}

func (t *ConvertedRewardPointRepositoryImpl) Update(ctx context.Context, data entity.ConvertedRewardPoint) error {
	if err := t.db.Model(entity.ConvertedRewardPoint{}).Where("rew_convert_id", data.RewConvertId).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (t *ConvertedRewardPointRepositoryImpl) GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*entity.ResponseEntity, error) {
	convertedRewardPoints := []entity.ConvertedRewardPoint{}
	limit, err := strconv.ParseInt(perPage, 10, 64)
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if err != nil {
		return nil, err
	}
	isFirstPage := cursor == ""
	pointsNext := false

	query := t.db

	if cursor != "" {
		decodedCursor, err := helpers.DecodeCursor(cursor)
		if err != nil {
			return nil, err
		}
		pointsNext = decodedCursor["points_next"] == true

		operator, order := t.getPaginationOperator(pointsNext, sortOrder)
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
		convertedRewardPoints = t.reverse(convertedRewardPoints)
	}

	pageInfo := t.calculatePagination(isFirstPage, hasPagination, int(limit), convertedRewardPoints, pointsNext)

	response := entity.ResponseEntity{
		Success:    true,
		Data:       convertedRewardPoints,
		Pagination: pageInfo,
	}

	return &response, nil
}

func (t *ConvertedRewardPointRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.ConvertedRewardPoint, error) {
	var resp entity.ConvertedRewardPoint

	tx := t.db.Model(entity.ConvertedRewardPoint{RewConvertId: id}).First(&resp)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &resp, nil
}

func (t *ConvertedRewardPointRepositoryImpl) getPaginationOperator(pointsNext bool, sortOrder string) (string, string) {
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

func (t *ConvertedRewardPointRepositoryImpl) calculatePagination(isFirstPage bool, hasPagination bool, limit int, convertedRewardPoints []entity.ConvertedRewardPoint, pointsNext bool) helpers.PaginationInfo {
	pagination := helpers.PaginationInfo{}
	nextCur := helpers.Cursor{}
	prevCur := helpers.Cursor{}
	if isFirstPage {
		if hasPagination {
			nextCur := helpers.CreateCursor(convertedRewardPoints[limit-1].RewConvertId, convertedRewardPoints[limit-1].CreatedAt, true)
			pagination = helpers.GeneratePager(nextCur, nil)
		}
	} else {
		if pointsNext {
			// if pointing next, it always has prev but it might not have next
			if hasPagination {
				nextCur = helpers.CreateCursor(convertedRewardPoints[limit-1].RewConvertId, convertedRewardPoints[limit-1].CreatedAt, true)
			}
			prevCur = helpers.CreateCursor(convertedRewardPoints[0].RewConvertId, convertedRewardPoints[0].CreatedAt, false)
			pagination = helpers.GeneratePager(nextCur, prevCur)
		} else {
			// this is case of prev, there will always be nest, but prev needs to be calculated
			nextCur = helpers.CreateCursor(convertedRewardPoints[limit-1].RewConvertId, convertedRewardPoints[limit-1].CreatedAt, true)
			if hasPagination {
				prevCur = helpers.CreateCursor(convertedRewardPoints[0].RewConvertId, convertedRewardPoints[0].CreatedAt, false)
			}
			pagination = helpers.GeneratePager(nextCur, prevCur)
		}
	}
	return pagination
}

func (t *ConvertedRewardPointRepositoryImpl) reverse(s []entity.ConvertedRewardPoint) []entity.ConvertedRewardPoint {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
