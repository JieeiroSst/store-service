package http

import (
	"context"
	"net/http"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
)

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	var in port.BlockInput
	if !decode(w, r, &in) {
		return
	}
	res, err := h.fraud.Block(r.Context(), actor, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	if err := h.fraud.Unblock(r.Context(), actor, r.PathValue("userId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	limit, offset, ok := paging(w, r)
	if !ok {
		return
	}
	p, err := h.fraud.Blocks(r.Context(), actor, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, p)
}

func (h *Handler) ListRewards(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	limit, offset, ok := paging(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	p, err := h.fraud.Rewards(r.Context(), actor, port.ListRewardsInput{Status: q.Get("status"), UserID: q.Get("user_id"), Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, p)
}

func (h *Handler) ApproveReward(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, h.fraud.Approve)
}

func (h *Handler) RejectReward(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, h.fraud.Reject)
}

func (h *Handler) review(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actor model.Actor, id int64) (*model.ReviewReward, error)) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	res, err := fn(r.Context(), actor, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

func (h *Handler) InvestigatePlays(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	limit, offset, ok := paging(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	p, err := h.fraud.Plays(r.Context(), actor, port.ListPlaysInput{UserID: q.Get("user_id"), DeviceID: q.Get("device_id"), IP: q.Get("ip"), Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, p)
}
