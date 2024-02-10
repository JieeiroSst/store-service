package repository

import "github.com/JIeeiroSst/search-service/internal"

type indexedTask struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Priority    internal.Priority `json:"priority"`
	IsDone      bool              `json:"is_done"`
	DateStart   int64             `json:"date_start"`
	DateDue     int64             `json:"date_due"`
}
