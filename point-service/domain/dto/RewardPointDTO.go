package dto

import "github.com/JIeeiroSst/point-service/domain/entity"

type RewardPointDTO struct {
	RewardPointsId  string `json:"reward_points_id"`
	TotalPoints     int    `json:"total_points"`
	PointsPending   int    `json:"points_pending"`
	PointsActive    int    `json:"points_active"`
	PointsExpired   int    `json:"points_expired"`
	PointsConverted int    `json:"points_converted"`
	PointsCancelled int    `json:"points_cancelled"`
	ActivateDate    int    `json:"activate_date"`
	ExpireDate      int    `json:"expire_date"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (g *RewardPointDTO) TransformListEntityToDto(f []entity.RewardPoint) []RewardPointDTO {
	var result []RewardPointDTO
	for _, fd := range f {
		result = append(result, RewardPointDTO{
			RewardPointsId:  fd.RewardPointsId,
			TotalPoints:     fd.TotalPoints,
			PointsPending:   fd.PointsPending,
			PointsActive:    fd.PointsActive,
			PointsExpired:   fd.PointsExpired,
			PointsConverted: fd.PointsConverted,
			PointsCancelled: fd.PointsCancelled,
			ActivateDate:    fd.ActivateDate,
			ExpireDate:      fd.ExpireDate,
		})
	}
	return result
}

func (g *RewardPointDTO) TransformEntityToDto(fd entity.RewardPoint) RewardPointDTO {
	var result RewardPointDTO
	result = RewardPointDTO{
		RewardPointsId:  fd.RewardPointsId,
		TotalPoints:     fd.TotalPoints,
		PointsPending:   fd.PointsPending,
		PointsActive:    fd.PointsActive,
		PointsExpired:   fd.PointsExpired,
		PointsConverted: fd.PointsConverted,
		PointsCancelled: fd.PointsCancelled,
		ActivateDate:    fd.ActivateDate,
		ExpireDate:      fd.ExpireDate,
	}

	return result
}

func (g *RewardPointDTO) TransformDTOtoEntity(fd RewardPointDTO) entity.RewardPoint {
	var result entity.RewardPoint
	result = entity.RewardPoint{
		RewardPointsId:  fd.RewardPointsId,
		TotalPoints:     fd.TotalPoints,
		PointsPending:   fd.PointsPending,
		PointsActive:    fd.PointsActive,
		PointsExpired:   fd.PointsExpired,
		PointsConverted: fd.PointsConverted,
		PointsCancelled: fd.PointsCancelled,
		ActivateDate:    fd.ActivateDate,
		ExpireDate:      fd.ExpireDate,
	}

	return result
}
