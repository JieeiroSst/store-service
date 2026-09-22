package handler

import (
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"go.uber.org/fx"
)

func newPartnerHandler(svc *services.PartnerService) *PartnerHandler {
	return NewPartnerHandler(*svc)
}

func newPartnershipHandler(svc *services.PartnershipService) *PartnershipHandler {
	return NewPartnershipHandler(*svc)
}

func newPartnershipsPartnerHandler(svc *services.PartnershipsPartnerService) *PartnershipsPartnerHandler {
	return NewPartnershipsPartnerHandler(*svc)
}

func newProjectHandler(svc *services.ProjectService) *ProjectHandler {
	return NewProjectHandler(*svc)
}

var Module = fx.Options(
	fx.Provide(newPartnerHandler),
	fx.Provide(newPartnershipHandler),
	fx.Provide(newPartnershipsPartnerHandler),
	fx.Provide(newProjectHandler),
)
