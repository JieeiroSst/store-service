package dto

import "github.com/JIeeiroSst/point-service/domain/entity"

type RewardDiscountDTO struct {
	RewardDiscountID string `json:"reward_discount_id"`
	TotalPoints      int    `json:"total_points"`
	PointsPending    int    `json:"points_pending"`
	PointsActive     int    `json:"points_active"`
	PointsExpired    int    `json:"points_expired"`
	PointsConverted  int    `json:"points_converted"`
	PointsCancelled  int    `json:"points_cancelled"`
	ActivateDate     int    `json:"activate_date"`
	ExpireDate       int    `json:"expire_date"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func (g *RewardDiscountDTO) TransformListEntityToDto(f []entity.RewardDiscount) []RewardDiscountDTO {
	var result []RewardDiscountDTO
	for _, fd := range f {
		result = append(result, RewardDiscountDTO{
			RewardDiscountID: fd.RewardDiscountID,
			TotalPoints:      fd.TotalPoints,
			PointsPending:    fd.PointsPending,
			PointsActive:     fd.PointsActive,
			PointsExpired:    fd.PointsExpired,
			PointsConverted:  fd.PointsConverted,
			PointsCancelled:  fd.PointsCancelled,
			ActivateDate:     fd.ActivateDate,
			ExpireDate:       fd.ExpireDate,
		})
	}
	return result
}

func (g *RewardDiscountDTO) TransformEntityToDto(fd entity.RewardDiscount) RewardDiscountDTO {
	var result RewardDiscountDTO
	result = RewardDiscountDTO{
		RewardDiscountID: fd.RewardDiscountID,
		TotalPoints:      fd.TotalPoints,
		PointsPending:    fd.PointsPending,
		PointsActive:     fd.PointsActive,
		PointsExpired:    fd.PointsExpired,
		PointsConverted:  fd.PointsConverted,
		PointsCancelled:  fd.PointsCancelled,
		ActivateDate:     fd.ActivateDate,
		ExpireDate:       fd.ExpireDate,
	}

	return result
}

func (g *RewardDiscountDTO) TransformDTOtoEntity(fd RewardDiscountDTO) entity.RewardDiscount {
	var result entity.RewardDiscount
	result = entity.RewardDiscount{
		RewardDiscountID: fd.RewardDiscountID,
		TotalPoints:      fd.TotalPoints,
		PointsPending:    fd.PointsPending,
		PointsActive:     fd.PointsActive,
		PointsExpired:    fd.PointsExpired,
		PointsConverted:  fd.PointsConverted,
		PointsCancelled:  fd.PointsCancelled,
		ActivateDate:     fd.ActivateDate,
		ExpireDate:       fd.ExpireDate,
	}

	return result
}
