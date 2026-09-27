package handler

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
)

const maxExportRows = 50000

type exportRequest struct {
	Sheet string `json:"sheet"`

	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows" binding:"required"`
}

func (h *Handler) exportXLSX(c *gin.Context) {
	var req exportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if len(req.Rows) == 0 || len(req.Rows) > maxExportRows {
		badRequest(c, fmt.Errorf("rows must contain 1..%d items", maxExportRows))
		return
	}
	cols := req.Columns
	if len(cols) == 0 {
		seen := map[string]bool{}
		for _, row := range req.Rows {
			for k := range row {
				if !seen[k] {
					seen[k] = true
					cols = append(cols, k)
				}
			}
		}
		sort.Strings(cols)
	}

	sheet := orDefault(req.Sheet, "Sheet1")
	x := excelize.NewFile()
	x.SetSheetName(x.GetSheetName(1), sheet)
	for i, col := range cols {
		x.SetCellValue(sheet, fmt.Sprintf("%s1", excelize.ToAlphaString(i)), col)
	}
	for r, row := range req.Rows {
		for i, col := range cols {
			if v, ok := row[col]; ok {
				x.SetCellValue(sheet, fmt.Sprintf("%s%d", excelize.ToAlphaString(i), r+2), v)
			}
		}
	}
	buf, err := x.WriteToBuffer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=data.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
