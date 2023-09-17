package dto

import "github.com/JIeeiroSst/point-service/domain/entity"

type ConvertedRewardPointDTO struct {
	RewConvertId          string `json:"rew_convert_id"`
	RewConvertOrdDetailId string `json:"rew_convert_ord_detail_id"` //FK
	RewConvertDiscountId  string `json:"rew_convert_discount_id"`   //FK
	RewConvertPoints      int    `json:"rew_convert_points"`
	RewConvertDate        int    `json:"rew_convert_date"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

func (g *ConvertedRewardPointDTO) TransformListEntityToDto(f []entity.ConvertedRewardPoint) []ConvertedRewardPointDTO {
	var result []ConvertedRewardPointDTO
	for _, fd := range f {
		result = append(result, ConvertedRewardPointDTO{
			RewConvertId:          fd.RewConvertId,
			RewConvertOrdDetailId: fd.RewConvertOrdDetailId,
			RewConvertDiscountId:  fd.RewConvertDiscountId,
			RewConvertPoints:      fd.RewConvertPoints,
			RewConvertDate:        fd.RewConvertDate,
		})
	}
	return result
}

func (g *ConvertedRewardPointDTO) TransformEntityToDto(fd entity.ConvertedRewardPoint) ConvertedRewardPointDTO {
	var result ConvertedRewardPointDTO
	result = ConvertedRewardPointDTO{
		RewConvertId:          fd.RewConvertId,
		RewConvertOrdDetailId: fd.RewConvertOrdDetailId,
		RewConvertDiscountId:  fd.RewConvertDiscountId,
		RewConvertPoints:      fd.RewConvertPoints,
		RewConvertDate:        fd.RewConvertDate,
	}
	return result
}

func (g *ConvertedRewardPointDTO) TransformDTOtoEntity(fd ConvertedRewardPointDTO) entity.ConvertedRewardPoint {
	var result entity.ConvertedRewardPoint
	result = entity.ConvertedRewardPoint{
		RewConvertId:          fd.RewConvertId,
		RewConvertOrdDetailId: fd.RewConvertOrdDetailId,
		RewConvertDiscountId:  fd.RewConvertDiscountId,
		RewConvertPoints:      fd.RewConvertPoints,
		RewConvertDate:        fd.RewConvertDate,
	}
	return result
}
