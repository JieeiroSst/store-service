package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, h *Handler) {
	r.HandleFunc("/healthz", h.Health).Methods(http.MethodGet)

	r.HandleFunc("/campaigns", h.CreateCampaign).Methods(http.MethodPost)
	r.HandleFunc("/campaigns", h.ListCampaigns).Methods(http.MethodGet)
	r.HandleFunc("/campaigns/{id}", h.GetCampaign).Methods(http.MethodGet)
	r.HandleFunc("/campaigns/{id}", h.UpdateCampaign).Methods(http.MethodPut)
	r.HandleFunc("/campaigns/{id}", h.DeleteCampaign).Methods(http.MethodDelete)
	r.HandleFunc("/campaigns/{id}/status", h.ChangeCampaignStatus).Methods(http.MethodPatch)
	r.HandleFunc("/campaigns/{id}/ads", h.ListAdsByCampaign).Methods(http.MethodGet)
	r.HandleFunc("/campaigns/{id}/performance", h.CampaignPerformance).Methods(http.MethodGet)

	r.HandleFunc("/categories", h.CreateCategory).Methods(http.MethodPost)
	r.HandleFunc("/categories", h.ListCategories).Methods(http.MethodGet)
	r.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
	r.HandleFunc("/categories/{id}", h.UpdateCategory).Methods(http.MethodPut)
	r.HandleFunc("/categories/{id}", h.DeleteCategory).Methods(http.MethodDelete)

	r.HandleFunc("/positions", h.CreatePosition).Methods(http.MethodPost)
	r.HandleFunc("/positions", h.ListPositions).Methods(http.MethodGet)
	r.HandleFunc("/positions/{id}", h.GetPosition).Methods(http.MethodGet)
	r.HandleFunc("/positions/{id}", h.UpdatePosition).Methods(http.MethodPut)
	r.HandleFunc("/positions/{id}", h.DeletePosition).Methods(http.MethodDelete)
	r.HandleFunc("/positions/{id}/serve", h.ServeAd).Methods(http.MethodGet)

	r.HandleFunc("/ads", h.CreateAd).Methods(http.MethodPost)
	r.HandleFunc("/ads", h.ListAds).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}", h.GetAd).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}", h.UpdateAd).Methods(http.MethodPut)
	r.HandleFunc("/ads/{id}", h.DeleteAd).Methods(http.MethodDelete)
	r.HandleFunc("/ads/{id}/impressions", h.ListImpressionsByAd).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}/clicks", h.ListClicksByAd).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}/performance/recompute", h.RecomputePerformance).Methods(http.MethodPost)

	r.HandleFunc("/position-mappings", h.CreatePositionMapping).Methods(http.MethodPost)
	r.HandleFunc("/position-mappings", h.ListPositionMappings).Methods(http.MethodGet)
	r.HandleFunc("/position-mappings/{id}", h.GetPositionMapping).Methods(http.MethodGet)
	r.HandleFunc("/position-mappings/{id}", h.UpdatePositionMapping).Methods(http.MethodPut)
	r.HandleFunc("/position-mappings/{id}", h.DeletePositionMapping).Methods(http.MethodDelete)

	r.HandleFunc("/impressions", h.ListImpressions).Methods(http.MethodGet)
	r.HandleFunc("/impressions/{id}", h.GetImpression).Methods(http.MethodGet)

	r.HandleFunc("/clicks", h.ListClicks).Methods(http.MethodGet)
	r.HandleFunc("/clicks/track", h.TrackClick).Methods(http.MethodGet)
	r.HandleFunc("/clicks/{id}", h.GetClick).Methods(http.MethodGet)

	r.HandleFunc("/targeting-rules", h.CreateTargetingRule).Methods(http.MethodPost)
	r.HandleFunc("/targeting-rules", h.ListTargetingRules).Methods(http.MethodGet)
	r.HandleFunc("/targeting-rules/{id}", h.GetTargetingRule).Methods(http.MethodGet)
	r.HandleFunc("/targeting-rules/{id}", h.UpdateTargetingRule).Methods(http.MethodPut)
	r.HandleFunc("/targeting-rules/{id}", h.DeleteTargetingRule).Methods(http.MethodDelete)

	r.HandleFunc("/performance-summaries", h.ListPerformanceSummaries).Methods(http.MethodGet)
	r.HandleFunc("/performance-summaries/{id}", h.GetPerformanceSummary).Methods(http.MethodGet)
}
