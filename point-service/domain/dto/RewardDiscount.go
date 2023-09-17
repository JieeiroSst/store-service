package dto

import "github.com/JIeeiroSst/point-service/domain/entity"

type RewardDiscountDRO struct {
	RewardDiscountID int    `json:"reward_discount_id"`
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

func (g *RewardDiscountDRO) TransformListEntityToDto(f []entity.RewardDiscount) []RewardDiscountDRO {
	var result []RewardDiscountDRO
	for _, fd := range f {
		result = append(result, RewardDiscountDRO{
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

func (g *RewardDiscountDRO) TransformEntityToDto(fd entity.RewardDiscount) RewardDiscountDRO {
	var result RewardDiscountDRO
	result = RewardDiscountDRO{
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
