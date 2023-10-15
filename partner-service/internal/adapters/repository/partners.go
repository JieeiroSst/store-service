package repository

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

func (m *DB) CreatePartnership(userID string, Partnership domain.Partnership) error

func (m *DB) ReadPartnership(id string) (*domain.Partnership, error)

func (m *DB) ReadPartnerships(pagination domain.Pagination) (*domain.Pagination, error)

func (m *DB) UpdatePartnership(id string, Partnership domain.Partnership) error

func (m *DB) DeletePartnership(id string) error
