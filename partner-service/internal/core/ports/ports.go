package ports

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

type PartnershipRepository interface {
	CreatePartnership(userID string, Partnership domain.Partnership) error
	ReadPartnership(id string) (*domain.Partnership, error)
	ReadPartnerships(pagination domain.Pagination) (*domain.Pagination, error)
	UpdatePartnership(id string, Partnership domain.Partnership) error
	DeletePartnership(id string) error
}

type PartnershipService interface {
	CreatePartnership(userID string, Partnership domain.Partnership) error
	ReadPartnership(id string) (*domain.Partnership, error)
	ReadPartnerships(pagination domain.Pagination) (*domain.Pagination, error)
	UpdatePartnership(id string, Partnership domain.Partnership) error
	DeletePartnership(id string) error
}

type PartnershipsPartnerRepository interface {
	CreatePartnershipsPartner(userID string, PartnershipsPartner domain.PartnershipsPartner) error
	ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error)
	ReadPartnershipsPartners(pagination domain.Pagination) (*domain.Pagination, error)
	UpdatePartnershipsPartner(id string, PartnershipsPartner domain.PartnershipsPartner) error
	DeletePartnershipsPartner(id string) error
}

type PartnershipsPartnerService interface {
	CreatePartnershipsPartner(userID string, PartnershipsPartner domain.PartnershipsPartner) error
	ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error)
	ReadPartnershipsPartners(pagination domain.Pagination) (*domain.Pagination, error)
	UpdatePartnershipsPartner(id string, PartnershipsPartner domain.PartnershipsPartner) error
	DeletePartnershipsPartner(id string) error
}

type PartnerRepository interface {
	CreatePartner(userID string, Partner domain.Partner) error
	ReadPartner(id string) (*domain.Partner, error)
	ReadPartners(pagination domain.Pagination) ([]*domain.Partner, error)
	UpdatePartner(id string, Partner domain.Partner) error
	DeletePartner(id string) error
}

type PartnerService interface {
	CreatePartner(userID string, Partner domain.Partner) error
	ReadPartner(id string) (*domain.Partner, error)
	ReadPartners(pagination domain.Pagination) (*domain.Pagination, error)
	UpdatePartner(id string, Partner domain.Partner) error
	DeletePartner(id string) error
}

type ProjectRepository interface {
	CreateProject(userID string, Project domain.Project) error
	ReadProject(id string) (*domain.Project, error)
	ReadProjects(pagination domain.Pagination) (*domain.Pagination, error)
	UpdateProject(id string, Project domain.Project) error
	DeleteProject(id string) error
}

type ProjectService interface {
	CreateProject(userID string, Project domain.Project) error
	ReadProject(id string) (*domain.Project, error)
	ReadProjects(pagination domain.Pagination) (*domain.Pagination, error)
	UpdateProject(id string, Project domain.Project) error
	DeleteProject(id string) error
}
