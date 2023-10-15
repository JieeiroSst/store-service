package repository

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

func (m *DB) CreatePartnershipsPartner(userID string, PartnershipsPartner domain.PartnershipsPartner) error

func (m *DB) ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error)

func (m *DB) ReadPartnershipsPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	var partnershipsPartners []*domain.PartnershipsPartner

	m.db.Scopes(paginate(partnershipsPartners, &pagination, m.db)).Find(&partnershipsPartners)
	pagination.Rows = partnershipsPartners

	return &pagination, nil
}

func (m *DB) UpdatePartnershipsPartner(id string, PartnershipsPartner domain.PartnershipsPartner) error

func (m *DB) DeletePartnershipsPartner(id string) error
