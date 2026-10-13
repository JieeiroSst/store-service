package http

import (
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/pkg/authn"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, h *Handler) {
	r.Use(authn.FromEnv().OptionalMiddleware)
	r.HandleFunc("/healthz", h.Health).Methods(http.MethodGet)

	r.HandleFunc("/campaigns", admin(h.CreateCampaign)).Methods(http.MethodPost)
	r.HandleFunc("/campaigns", admin(h.ListCampaigns)).Methods(http.MethodGet)
	r.HandleFunc("/campaigns/{id}", admin(h.GetCampaign)).Methods(http.MethodGet)
	r.HandleFunc("/campaigns/{id}", admin(h.UpdateCampaign)).Methods(http.MethodPut)
	r.HandleFunc("/campaigns/{id}", admin(h.DeleteCampaign)).Methods(http.MethodDelete)
	r.HandleFunc("/campaigns/{id}/status", admin(h.ChangeCampaignStatus)).Methods(http.MethodPatch)
	r.HandleFunc("/campaigns/{id}/ads", admin(h.ListAdsByCampaign)).Methods(http.MethodGet)
	r.HandleFunc("/campaigns/{id}/performance", admin(h.CampaignPerformance)).Methods(http.MethodGet)

	r.HandleFunc("/categories", admin(h.CreateCategory)).Methods(http.MethodPost)
	r.HandleFunc("/categories", admin(h.ListCategories)).Methods(http.MethodGet)
	r.HandleFunc("/categories/{id}", admin(h.GetCategory)).Methods(http.MethodGet)
	r.HandleFunc("/categories/{id}", admin(h.UpdateCategory)).Methods(http.MethodPut)
	r.HandleFunc("/categories/{id}", admin(h.DeleteCategory)).Methods(http.MethodDelete)

	r.HandleFunc("/positions", admin(h.CreatePosition)).Methods(http.MethodPost)
	r.HandleFunc("/positions", admin(h.ListPositions)).Methods(http.MethodGet)
	r.HandleFunc("/positions/{id}", admin(h.GetPosition)).Methods(http.MethodGet)
	r.HandleFunc("/positions/{id}", admin(h.UpdatePosition)).Methods(http.MethodPut)
	r.HandleFunc("/positions/{id}", admin(h.DeletePosition)).Methods(http.MethodDelete)
	r.HandleFunc("/positions/{id}/serve", h.ServeAd).Methods(http.MethodGet)

	r.HandleFunc("/ads", admin(h.CreateAd)).Methods(http.MethodPost)
	r.HandleFunc("/ads", admin(h.ListAds)).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}", admin(h.GetAd)).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}", admin(h.UpdateAd)).Methods(http.MethodPut)
	r.HandleFunc("/ads/{id}", admin(h.DeleteAd)).Methods(http.MethodDelete)
	r.HandleFunc("/ads/{id}/impressions", admin(h.ListImpressionsByAd)).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}/clicks", admin(h.ListClicksByAd)).Methods(http.MethodGet)
	r.HandleFunc("/ads/{id}/performance/recompute", admin(h.RecomputePerformance)).Methods(http.MethodPost)

	r.HandleFunc("/position-mappings", admin(h.CreatePositionMapping)).Methods(http.MethodPost)
	r.HandleFunc("/position-mappings", admin(h.ListPositionMappings)).Methods(http.MethodGet)
	r.HandleFunc("/position-mappings/{id}", admin(h.GetPositionMapping)).Methods(http.MethodGet)
	r.HandleFunc("/position-mappings/{id}", admin(h.UpdatePositionMapping)).Methods(http.MethodPut)
	r.HandleFunc("/position-mappings/{id}", admin(h.DeletePositionMapping)).Methods(http.MethodDelete)

	r.HandleFunc("/impressions", admin(h.ListImpressions)).Methods(http.MethodGet)
	r.HandleFunc("/impressions/{id}", admin(h.GetImpression)).Methods(http.MethodGet)

	r.HandleFunc("/clicks", admin(h.ListClicks)).Methods(http.MethodGet)
	r.HandleFunc("/clicks/track", h.TrackClick).Methods(http.MethodGet)
	r.HandleFunc("/clicks/{id}", admin(h.GetClick)).Methods(http.MethodGet)

	r.HandleFunc("/targeting-rules", admin(h.CreateTargetingRule)).Methods(http.MethodPost)
	r.HandleFunc("/targeting-rules", admin(h.ListTargetingRules)).Methods(http.MethodGet)
	r.HandleFunc("/targeting-rules/{id}", admin(h.GetTargetingRule)).Methods(http.MethodGet)
	r.HandleFunc("/targeting-rules/{id}", admin(h.UpdateTargetingRule)).Methods(http.MethodPut)
	r.HandleFunc("/targeting-rules/{id}", admin(h.DeleteTargetingRule)).Methods(http.MethodDelete)

	r.HandleFunc("/performance-summaries", admin(h.ListPerformanceSummaries)).Methods(http.MethodGet)
	r.HandleFunc("/performance-summaries/{id}", admin(h.GetPerformanceSummary)).Methods(http.MethodGet)
}
