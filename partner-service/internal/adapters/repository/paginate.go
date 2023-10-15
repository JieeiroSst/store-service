package repository

import (
	"math"
	"strings"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"gorm.io/gorm"
)

func paginate(value interface{}, preload string, pagination *domain.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		if preload != "" {
			preloads := strings.Split(preload, ",")
			if len(preloads) == 2 {
				return db.Preload(preloads[0]).Preload(preloads[1]).Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
			}
			return db.Preload(preloads[0]).Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
		}
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}
