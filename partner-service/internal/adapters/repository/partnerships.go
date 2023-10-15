package repository

import (
	"errors"
	"fmt"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
)

func (m *DB) CreatePartnership(userID string, partnership domain.Partnership) error {
	partnership.ID = snowflakeID()
	partnership.UserID = userID
	req := m.db.Create(&partnership)
	if req.RowsAffected == 0 {
		return fmt.Errorf("partnership not saved: %v", req.Error)
	}
	return nil
}

func (m *DB) ReadPartnership(id string) (*domain.Partnership, error) {
	partner := &domain.Partnership{}
	req := m.db.Preload("Projects").First(&partner, "id = ? ", id)
	if req.RowsAffected == 0 {
		return nil, errors.New("partner not found")
	}
	return partner, nil
}

func (m *DB) ReadPartnerships(pagination domain.Pagination) (*domain.Pagination, error) {
	var partnership []*domain.Partnership

	m.db.Scopes(paginate(partnership, "Projects", &pagination, m.db)).Find(&partnership)
	pagination.Rows = partnership

	return &pagination, nil
}

func (m *DB) UpdatePartnership(id string, partnership domain.Partnership) error {
	req := m.db.Model(&partnership).Where("id = ?", id).Updates(partnership)
	if req.RowsAffected == 0 {
		return errors.New("partnership not found")
	}
	return nil
}

func (m *DB) DeletePartnership(id string) error {
	partner := &domain.Partnership{}
	req := m.db.Where("id = ?", id).Delete(&partner)
	if req.RowsAffected == 0 {
		return errors.New("partnership not found")
	}
	return nil
}
