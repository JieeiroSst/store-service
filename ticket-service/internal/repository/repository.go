package repository

import "gorm.io/gorm"

type Repository struct {
	Customers
	Invoices
	Tickets
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		Customers: NewCustomerRepository(db),
		Invoices:  NewInvoicesRepository(db),
		Tickets:   NewTicketsRepository(db),
	}
}
