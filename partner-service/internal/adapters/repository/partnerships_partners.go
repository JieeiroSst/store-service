package repository

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

func (m *DB) CreatePartnershipsPartner(userID string, PartnershipsPartner domain.PartnershipsPartner) error {

	return nil
}

func (m *DB) ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error) {
	return nil, nil
}

func (m *DB) ReadPartnershipsPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	var partnershipsPartners []*domain.PartnershipsPartner

	m.db.Scopes(paginate(partnershipsPartners,"Partners,Partnerships", &pagination, m.db)).Find(&partnershipsPartners)
	pagination.Rows = partnershipsPartners

	return &pagination, nil
}

func (m *DB) UpdatePartnershipsPartner(id string, PartnershipsPartner domain.PartnershipsPartner) error {
	return nil
}

func (m *DB) DeletePartnershipsPartner(id string) error {
	return nil
}
