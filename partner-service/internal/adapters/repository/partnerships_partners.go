package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
)

func (m *DB) CreatePartnershipsPartner(userID string, partnershipsPartner domain.PartnershipsPartner) error {
	partnershipsPartner.ID = snowflakeID()
	partnershipsPartner.UserID = userID
	partnershipsPartner.CreatedAt = int(time.Now().Unix())
	req := m.db.Create(&partnershipsPartner)
	if req.RowsAffected == 0 {
		return fmt.Errorf("partnershipsPartner not saved: %v", req.Error)
	}
	return nil
}

func (m *DB) ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error) {
	partnershipsPartner := &domain.PartnershipsPartner{}
	req := m.db.Preload("Partners").Preload("Partnerships").First(&partnershipsPartner, "id = ? ", id)
	if req.RowsAffected == 0 {
		return nil, errors.New("partner not found")
	}
	return partnershipsPartner, nil
}

func (m *DB) ReadPartnershipsPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	var partnershipsPartners []*domain.PartnershipsPartner

	m.db.Scopes(paginate(partnershipsPartners, "Partners,Partnerships", &pagination, m.db)).Find(&partnershipsPartners)
	pagination.Rows = partnershipsPartners

	return &pagination, nil
}

func (m *DB) UpdatePartnershipsPartner(id string, partnershipsPartner domain.PartnershipsPartner) error {
	partnershipsPartner.CreatedAt = int(time.Now().Unix())
	req := m.db.Model(&partnershipsPartner).Where("id = ?", id).Updates(partnershipsPartner)
	if req.RowsAffected == 0 {
		return errors.New("partnershipsPartner not found")
	}
	return nil
}

func (m *DB) DeletePartnershipsPartner(id string) error {
	partnershipsPartner := &domain.PartnershipsPartner{}
	req := m.db.Where("id = ?", id).Delete(&partnershipsPartner)
	if req.RowsAffected == 0 {
		return errors.New("partnershipsPartner not found")
	}
	return nil
}
