package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

func (h *Handler) ServeAd(w http.ResponseWriter, r *http.Request) {
	positionID, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	q := r.URL.Query()
	target := model.TargetContext{
		SessionID:   q.Get("session_id"),
		Country:     q.Get("country"),
		Device:      q.Get("device"),
		Gender:      q.Get("gender"),
		ReferrerURL: q.Get("referrer_url"),
		PageURL:     q.Get("page_url"),
		UserAgent:   r.UserAgent(),
		IPAddress:   clientIP(r),
		HourOfDay:   time.Now().Hour(),
	}
	if age, err := strconv.Atoi(q.Get("age")); err == nil {
		target.Age = age
	}
	if uid, err := strconv.ParseUint(q.Get("user_id"), 10, 64); err == nil {
		u := uint(uid)
		target.UserID = &u
	}

	served, err := h.serving.Serve(r.Context(), positionID, target)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, served)
}

func (h *Handler) TrackClick(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	adID, err := strconv.ParseUint(q.Get("ad_id"), 10, 64)
	if err != nil {
		writeError(w, common.ErrInvalidInput)
		return
	}

	var impressionID uint64
	if raw := q.Get("impression_id"); raw != "" {
		impressionID, _ = strconv.ParseUint(raw, 10, 64)
	}

	target := model.TargetContext{
		SessionID:   q.Get("session_id"),
		ReferrerURL: q.Get("referrer_url"),
		UserAgent:   r.UserAgent(),
		IPAddress:   clientIP(r),
	}

	targetURL, err := h.serving.TrackClick(r.Context(), uint(adID), uint(impressionID), target)
	if err != nil {
		writeError(w, err)
		return
	}

	if targetURL == "" {
		writeJSON(w, http.StatusOK, map[string]string{"target_url": ""})
		return
	}
	http.Redirect(w, r, targetURL, http.StatusFound)
}
