package repository

import (
	"errors"
	"fmt"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
)

func (m *DB) CreatePartner(userID string, partner domain.Partner) error {
	partner = domain.Partner{
		ID:     snowflakeID(),
		Type:   partner.Type,
		UserID: userID,
	}
	req := m.db.Create(&partner)
	if req.RowsAffected == 0 {
		return fmt.Errorf("partner not saved: %v", req.Error)
	}
	return nil
}

func (m *DB) ReadPartner(id string) (*domain.Partner, error) {
	partner := &domain.Partner{}
	req := m.db.First(&partner, "id = ? ", id)
	if req.RowsAffected == 0 {
		return nil, errors.New("partner not found")
	}
	return partner, nil
}

func (m *DB) ReadPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	var partners []*domain.Partner

	m.db.Scopes(paginate(partners, "", &pagination, m.db)).Find(&partners)
	pagination.Rows = partners

	return &pagination, nil
}

func (m *DB) UpdatePartner(id string, partner domain.Partner) error {
	req := m.db.Model(&partner).Where("id = ?", id).Updates(partner)
	if req.RowsAffected == 0 {
		return errors.New("partner not found")
	}
	return nil
}

func (m *DB) DeletePartner(id string) error {
	partner := &domain.Partner{}
	req := m.db.Where("id = ?", id).Delete(&partner)
	if req.RowsAffected == 0 {
		return errors.New("partner not found")
	}
	return nil
}
