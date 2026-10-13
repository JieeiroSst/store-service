package http

import (
	"net/http"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
)

type moveLineRequest struct {
	TargetWishlistID int64 `json:"target_wishlist_id"`
}

func (h *Handler) DefaultWishlist(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	out, err := h.wishlists.Default(r.Context(), actor)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) AddToDefault(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	var in port.AddLineInput
	if !decode(w, r, &in) {
		return
	}
	out, err := h.wishlists.AddToDefault(r.Context(), actor, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, out)
}

func (h *Handler) FindProduct(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	productID, ok := pathID(w, r, "productId")
	if !ok {
		return
	}
	out, err := h.wishlists.FindProduct(r.Context(), actor, productID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"product_id": productID, "in_wishlist": len(out) > 0, "wishlists": out})
}

func (h *Handler) CopySharedWishlist(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	out, err := h.wishlists.CopyShared(r.Context(), actor, r.PathValue("key"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, out)
}

func (h *Handler) ClearLines(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	n, err := h.wishlists.ClearLines(r.Context(), actor, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]int64{"removed": n})
}

func (h *Handler) MoveLine(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	lineID, ok := pathID(w, r, "lineId")
	if !ok {
		return
	}
	var in moveLineRequest
	if !decode(w, r, &in) {
		return
	}
	out, err := h.wishlists.MoveLine(r.Context(), actor, id, lineID, in.TargetWishlistID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) PopularProducts(w http.ResponseWriter, r *http.Request) {
	limit, err := queryInt(r, "limit")
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := h.wishlists.PopularProducts(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}
