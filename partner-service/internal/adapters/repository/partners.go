package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
)

const partnerCacheTTL = 5 * time.Minute

const partnerInactivityThreshold = 90 * 24 * time.Hour

func partnerCacheKey(id string) string {
	return fmt.Sprintf("partner:%s", id)
}

func (m *DB) CreatePartner(userID string, partner domain.Partner) error {
	status := partner.Status
	if status == "" {
		status = domain.PartnerStatusActive
	}
	now := int(time.Now().Unix())
	partner = domain.Partner{
		ID:        snowflakeID(),
		Type:      partner.Type,
		Name:      partner.Name,
		Email:     partner.Email,
		Phone:     partner.Phone,
		Address:   partner.Address,
		Status:    status,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	req := m.db.Create(&partner)
	if req.RowsAffected == 0 {
		return fmt.Errorf("partner not saved: %v", req.Error)
	}
	return nil
}

func (m *DB) ReadPartner(id string) (*domain.Partner, error) {
	partner := &domain.Partner{}

	if m.cache != nil {
		if err := m.cache.Get(partnerCacheKey(id), partner); err == nil {
			return partner, nil
		}
	}

	req := m.db.First(&partner, "id = ? AND deleted_at = 0", id)
	if req.RowsAffected == 0 {
		return nil, errors.New("partner not found")
	}

	if m.cache != nil {
		_ = m.cache.Set(partnerCacheKey(id), partner, partnerCacheTTL)
	}

	return partner, nil
}

func (m *DB) ReadPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	var partners []*domain.Partner

	query := m.db.Model(&domain.Partner{}).Where("deleted_at = 0")
	query.Scopes(paginate(partners, "", &pagination, query)).Find(&partners)
	pagination.Rows = partners

	return &pagination, nil
}

func (m *DB) SearchPartners(filter domain.PartnerFilter, pagination domain.Pagination) (*domain.Pagination, error) {
	var partners []*domain.Partner

	query := m.db.Model(&domain.Partner{}).Where("deleted_at = 0")
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	query.Scopes(paginate(partners, "", &pagination, query)).Find(&partners)
	pagination.Rows = partners

	return &pagination, nil
}

func (m *DB) UpdatePartner(id string, partner domain.Partner) error {
	partner.UpdatedAt = int(time.Now().Unix())
	req := m.db.Model(&partner).Where("id = ? AND deleted_at = 0", id).Updates(partner)
	if req.RowsAffected == 0 {
		return errors.New("partner not found")
	}

	if m.cache != nil {
		_ = m.cache.Delete(partnerCacheKey(id))
	}
	return nil
}

func (m *DB) UpdatePartnerStatus(id string, status string) error {
	req := m.db.Model(&domain.Partner{}).Where("id = ? AND deleted_at = 0", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": int(time.Now().Unix()),
	})
	if req.RowsAffected == 0 {
		return errors.New("partner not found")
	}

	if m.cache != nil {
		_ = m.cache.Delete(partnerCacheKey(id))
	}
	return nil
}

func (m *DB) DeletePartner(id string) error {
	req := m.db.Model(&domain.Partner{}).Where("id = ? AND deleted_at = 0", id).Updates(map[string]interface{}{
		"deleted_at": int(time.Now().Unix()),
	})
	if req.RowsAffected == 0 {
		return errors.New("partner not found")
	}

	if m.cache != nil {
		_ = m.cache.Delete(partnerCacheKey(id))
	}
	return nil
}

func (m *DB) RestorePartner(id string) error {
	req := m.db.Model(&domain.Partner{}).Where("id = ? AND deleted_at <> 0", id).Updates(map[string]interface{}{
		"deleted_at": 0,
		"updated_at": int(time.Now().Unix()),
	})
	if req.RowsAffected == 0 {
		return errors.New("deleted partner not found")
	}

	if m.cache != nil {
		_ = m.cache.Delete(partnerCacheKey(id))
	}
	return nil
}

func (m *DB) GetPartnerStatistics() (*domain.PartnerStatistics, error) {
	stats := &domain.PartnerStatistics{
		ByStatus: make(map[string]int64),
		ByType:   make(map[string]int64),
	}

	var total int64
	if err := m.db.Model(&domain.Partner{}).Where("deleted_at = 0").Count(&total).Error; err != nil {
		return nil, err
	}
	stats.Total = total

	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := m.db.Model(&domain.Partner{}).
		Select("status, count(*) as count").
		Where("deleted_at = 0").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	var typeCounts []struct {
		Type  string
		Count int64
	}
	if err := m.db.Model(&domain.Partner{}).
		Select("type, count(*) as count").
		Where("deleted_at = 0").
		Group("type").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}
	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}

	return stats, nil
}

func (m *DB) CheckPartnerActivity(id string) (*domain.PartnerActivity, error) {
	partner := &domain.Partner{}
	req := m.db.First(partner, "id = ? AND deleted_at = 0", id)
	if req.RowsAffected == 0 {
		return nil, errors.New("partner not found")
	}

	var openCount int64
	if err := m.db.Model(&domain.PartnershipsPartner{}).
		Where("partner_id = ? AND left_on = 0", id).
		Count(&openCount).Error; err != nil {
		return nil, err
	}
	hasOpenPartnership := openCount > 0

	var lastActivity int
	row := m.db.Model(&domain.PartnershipsPartner{}).
		Select("COALESCE(MAX(GREATEST(joined_on, left_on)), 0)").
		Where("partner_id = ?", id).
		Row()
	if err := row.Scan(&lastActivity); err != nil {
		return nil, err
	}
	if lastActivity == 0 {
		lastActivity = partner.CreatedAt
	}

	withinThreshold := time.Since(time.Unix(int64(lastActivity), 0)) < partnerInactivityThreshold
	lowScore := partner.Score < domain.PartnerScoreThreshold
	eligibleForClosure := !hasOpenPartnership && !withinThreshold && lowScore
	isActive := partner.Status == domain.PartnerStatusActive && !eligibleForClosure

	return &domain.PartnerActivity{
		PartnerID:          id,
		Status:             partner.Status,
		Score:              partner.Score,
		HasOpenPartnership: hasOpenPartnership,
		LastActivityAt:     lastActivity,
		EligibleForClosure: eligibleForClosure,
		IsActive:           isActive,
	}, nil
}

func (m *DB) CloseInactivePartners() (int64, error) {
	cutoff := time.Now().Add(-partnerInactivityThreshold).Unix()
	now := int(time.Now().Unix())

	var closedIDs []string
	err := m.db.Raw(`
		UPDATE partners
		SET status = ?, updated_at = ?
		WHERE deleted_at = 0
		  AND status = ?
		  AND score < ?
		  AND NOT EXISTS (
		      SELECT 1 FROM partnerships_partners pp
		      WHERE pp.partner_id = partners.id AND pp.left_on = 0
		  )
		  AND COALESCE(
		      (SELECT MAX(GREATEST(pp2.joined_on, pp2.left_on)) FROM partnerships_partners pp2 WHERE pp2.partner_id = partners.id),
		      partners.created_at
		  ) < ?
		RETURNING id
	`, domain.PartnerStatusInactive, now, domain.PartnerStatusActive, domain.PartnerScoreThreshold, cutoff).Scan(&closedIDs).Error
	if err != nil {
		return 0, err
	}

	if m.cache != nil {
		for _, id := range closedIDs {
			_ = m.cache.Delete(partnerCacheKey(id))
		}
	}

	return int64(len(closedIDs)), nil
}

func (m *DB) AdjustPartnerScore(id string, delta int) (*domain.Partner, error) {
	now := int(time.Now().Unix())

	var partner domain.Partner
	err := m.db.Raw(`
		UPDATE partners
		SET score = LEAST(GREATEST(score + ?, 0), ?), updated_at = ?
		WHERE id = ? AND deleted_at = 0
		RETURNING *
	`, delta, domain.PartnerScoreMax, now, id).Scan(&partner).Error
	if err != nil {
		return nil, err
	}
	if partner.ID == "" {
		return nil, errors.New("partner not found")
	}

	if m.cache != nil {
		_ = m.cache.Delete(partnerCacheKey(id))
	}

	return &partner, nil
}
