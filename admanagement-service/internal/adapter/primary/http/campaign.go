package http

import (
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req model.AdCampaign
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.campaign.Create(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	campaign, err := h.campaign.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.AdCampaign
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	req.ID = id

	if err := h.campaign.Update(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.campaign.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.campaign.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaigns)
}
 
func (h *Handler) ChangeCampaignStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req struct {
		Status model.CampaignStatus `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	campaign, err := h.campaign.ChangeStatus(r.Context(), id, req.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *Handler) ListAdsByCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	ads, err := h.ad.ListByCampaign(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ads)
}

func (h *Handler) CampaignPerformance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	from, to := parseDateRange(r)

	rollup, err := h.performanceSummary.CampaignRollup(r.Context(), id, from, to)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rollup)
}
