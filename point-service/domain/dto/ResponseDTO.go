package dto

import (
	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/JIeeiroSst/point-service/helpers"
)

type ResponseDTO struct {
	Success    bool                   `json:"success"`
	Data       interface{}            `json:"data"`
	Pagination helpers.PaginationInfo `json:"pagination"`
}

func (g *ResponseDTO) TransformEntityToDto(fd entity.ResponseEntity) *ResponseDTO {
	return &ResponseDTO{
		Success:    fd.Success,
		Data:       fd.Data,
		Pagination: fd.Pagination,
	}
}

func (g *ResponseDTO) TransformDTOtoEntity() entity.ResponseEntity {
	return entity.ResponseEntity{
		Success:    g.Success,
		Data:       g.Data,
		Pagination: g.Pagination,
	}
}
