package repository

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
)

type campaignTargetSource struct {
	db *gorm.DB
}

func NewCampaignTargetSource(db *gorm.DB) port.CampaignTargetSource {
	return &campaignTargetSource{db: db}
}

func (s *campaignTargetSource) scope(ctx context.Context, c *model.Campaign) (*gorm.DB, string, error) {
	switch c.Audience {
	case model.AudienceAllDevices:
		return s.db.WithContext(ctx).Table("notification_user_device AS t").Where("t.is_active = ?", true), "t.id", nil
	case model.AudienceUsers:
		return s.db.WithContext(ctx).Table("notification_user_device AS t").
			Joins("JOIN notification_campaign_recipient r ON r.user_id = t.user_id AND r.campaign_id = ?", c.ID).
			Where("t.is_active = ?", true), "t.id", nil
	case model.AudienceEmails:
		return s.db.WithContext(ctx).Table("notification_campaign_recipient AS t").Where("t.campaign_id = ?", c.ID), "t.id", nil
	default:
		return nil, "", fmt.Errorf("audience %s has no recipient list", c.Audience)
	}
}

func (s *campaignTargetSource) NextIDs(ctx context.Context, c *model.Campaign, afterID uint, limit int) ([]uint, error) {
	q, idCol, err := s.scope(ctx, c)
	if err != nil {
		return nil, err
	}
	var ids []uint
	err = q.Where(idCol+" > ?", afterID).Order(idCol+" ASC").Limit(limit).Pluck(idCol, &ids).Error
	return ids, err
}

func (s *campaignTargetSource) Targets(ctx context.Context, c *model.Campaign, afterID, untilID uint) ([]model.BatchTarget, error) {
	q, idCol, err := s.scope(ctx, c)
	if err != nil {
		return nil, err
	}
	cols := idCol + " AS id, t.user_id AS user_id, t.device_type AS device_type, t.device_token AS token, '' AS email"
	if c.Audience == model.AudienceEmails {
		cols = idCol + " AS id, 0 AS user_id, '' AS device_type, '' AS token, t.email AS email"
	}
	var targets []model.BatchTarget
	err = q.Select(cols).Where(idCol+" > ? AND "+idCol+" <= ?", afterID, untilID).Order(idCol + " ASC").Scan(&targets).Error
	return targets, err
}
