package entity

import "github.com/JIeeiroSst/point-service/helpers"

type ResponseDTO struct {
	Success    bool                   `json:"success"`
	Data       interface{}            `json:"data"`
	Pagination helpers.PaginationInfo `json:"pagination"`
}
