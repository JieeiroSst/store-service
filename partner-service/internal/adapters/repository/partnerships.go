package repository

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

func (m *DB) CreatePartnership(userID string, Partnership domain.Partnership) error {
	return nil
}

func (m *DB) ReadPartnership(id string) (*domain.Partnership, error) {
	return nil, nil
}

func (m *DB) ReadPartnerships(pagination domain.Pagination) (*domain.Pagination, error) {
	var partnership []*domain.Partnership

	m.db.Scopes(paginate(partnership, &pagination, m.db)).Find(&partnership)
	pagination.Rows = partnership

	return &pagination, nil
}

func (m *DB) UpdatePartnership(id string, Partnership domain.Partnership) error {
	return nil
}

func (m *DB) DeletePartnership(id string) error {
	return nil
}
