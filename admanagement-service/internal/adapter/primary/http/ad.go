package http

import (
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

func (h *Handler) CreateAd(w http.ResponseWriter, r *http.Request) {
	var req model.Ad
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.ad.Create(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) GetAd(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	ad, err := h.ad.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ad)
}

func (h *Handler) UpdateAd(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.Ad
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	req.ID = id

	if err := h.ad.Update(r.Context(), &req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (h *Handler) DeleteAd(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.ad.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ListAds(w http.ResponseWriter, r *http.Request) {
	ads, err := h.ad.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ads)
}
