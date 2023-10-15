package repository

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

func (m *DB) CreatePartner(userID string, Partner domain.Partner) error

func (m *DB) ReadPartner(id string) (*domain.Partner, error)

func (m *DB) ReadPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	var partners []*domain.Partner

	m.db.Scopes(paginate(partners, &pagination, m.db)).Find(&partners)
	pagination.Rows = partners

	return &pagination, nil
}

func (m *DB) UpdatePartner(id string, Partner domain.Partner) error

func (m *DB) DeletePartner(id string) error
