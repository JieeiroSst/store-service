package repository

import (
	"context"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"gorm.io/gorm"
)

type reporter struct{ db *gorm.DB }

func NewReporter(db *gorm.DB) port.Reporter { return &reporter{db: db} }

func (r *reporter) Summary(ctx context.Context) (*model.Summary, error) {
	db := conn(ctx, r.db)
	s := &model.Summary{
		Counts:        map[string]int64{},
		LeadsByStatus: map[string]int64{},
	}

	tables := map[string]any{
		"campaigns": &model.Campaign{}, "leads": &model.Lead{}, "accounts": &model.Account{},
		"contacts": &model.Contact{}, "cases": &model.Case{}, "contracts": &model.Contract{},
		"opportunities": &model.Opportunity{},
	}
	for name, m := range tables {
		var n int64
		if err := db.Model(m).Count(&n).Error; err != nil {
			return nil, common.ErrDBFailed
		}
		s.Counts[name] = n
	}

	if err := db.Model(&model.Case{}).Where("status = ?", model.CaseOpen).Count(&s.OpenCases).Error; err != nil {
		return nil, common.ErrDBFailed
	}

	var leads []struct {
		Status string
		Count  int64
	}
	if err := db.Model(&model.Lead{}).Select("status, count(*) as count").Group("status").Scan(&leads).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	for _, l := range leads {
		s.LeadsByStatus[l.Status] = l.Count
	}

	var rows []model.StageSummary
	if err := db.Model(&model.Opportunity{}).
		Select("opportunity_stage as stage, count(*) as count, coalesce(sum(amount), 0) as amount").
		Group("opportunity_stage").Scan(&rows).Error; err != nil {
		return nil, common.ErrDBFailed
	}

	byStage := map[string]model.StageSummary{}
	for _, row := range rows {
		byStage[row.Stage] = row
	}
	s.Pipeline = make([]model.StageSummary, 0, len(model.Stages))
	for _, stage := range model.Stages {
		row := byStage[stage]
		row.Stage = stage
		s.Pipeline = append(s.Pipeline, row)
	}
	return s, nil
}
