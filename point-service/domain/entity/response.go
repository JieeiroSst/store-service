package entity

import "github.com/JIeeiroSst/point-service/helpers"

type ResponseEntity struct {
	Success    bool                   `json:"success"`
	Data       interface{}            `json:"data"`
	Pagination helpers.PaginationInfo `json:"pagination"`
}
