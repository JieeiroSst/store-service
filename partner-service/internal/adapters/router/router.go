package router

import (
	"github.com/JIeeiroSst/partner-service/internal/adapters/handler"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	partnerHandler *handler.PartnerHandler,
	partnershipHandler *handler.PartnershipHandler,
	partnershipsPartnerHandler *handler.PartnershipsPartnerHandler,
	projectHandler *handler.ProjectHandler,
) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/v1")

	pprof.Register(router)

	partnerGroup := v1.Group("/partner")
	{
		partnerGroup.GET("/", partnerHandler.ReadPartners)
		partnerGroup.GET("/search", partnerHandler.SearchPartners)
		partnerGroup.GET("/statistics", partnerHandler.GetPartnerStatistics)
		partnerGroup.GET("/active", partnerHandler.CheckPartnerActivity)
		partnerGroup.GET("/:id", partnerHandler.ReadPartner)
		partnerGroup.POST("/", partnerHandler.CreatePartner)
		partnerGroup.POST("/close-inactive", partnerHandler.CloseInactivePartners)
		partnerGroup.PUT("/", partnerHandler.UpdatePartner)
		partnerGroup.PATCH("/status", partnerHandler.UpdatePartnerStatus)
		partnerGroup.PATCH("/restore", partnerHandler.RestorePartner)
		partnerGroup.PATCH("/score", partnerHandler.AdjustPartnerScore)
		partnerGroup.DELETE("/", partnerHandler.DeletePartner)
	}

	partnershipGroup := v1.Group("/partnership")
	{
		partnershipGroup.POST("/", partnershipHandler.CreatePartnership)
		partnershipGroup.GET("/:id", partnershipHandler.ReadPartnership)
		partnershipGroup.GET("/", partnershipHandler.ReadPartnerships)
		partnershipGroup.PUT("/", partnershipHandler.UpdatePartnership)
		partnershipGroup.DELETE("/", partnershipHandler.DeletePartnership)
	}

	partnershipsPartnerGroup := v1.Group("/partnerships-partner")
	{
		partnershipsPartnerGroup.POST("/", partnershipsPartnerHandler.CreatePartnershipsPartner)
		partnershipsPartnerGroup.GET("/:id", partnershipsPartnerHandler.ReadPartnershipsPartner)
		partnershipsPartnerGroup.GET("/", partnershipsPartnerHandler.ReadPartnershipsPartners)
		partnershipsPartnerGroup.PUT("/", partnershipsPartnerHandler.UpdatePartnershipsPartner)
		partnershipsPartnerGroup.DELETE("/", partnershipsPartnerHandler.DeletePartnershipsPartner)
	}

	projectGroup := v1.Group("/project")
	{
		projectGroup.POST("/", projectHandler.CreateProject)
		projectGroup.GET("/:id", projectHandler.ReadProject)
		projectGroup.GET("/", projectHandler.ReadProjects)
		projectGroup.PUT("/", projectHandler.UpdateProject)
		projectGroup.DELETE("/", projectHandler.DeleteProject)
	}

	return router
}
