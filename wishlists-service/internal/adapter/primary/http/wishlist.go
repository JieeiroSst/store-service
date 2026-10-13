package http

import (
	"net/http"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
)

func (h *Handler) CreateWishlist(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	var in port.CreateWishlistInput
	if !decode(w, r, &in) {
		return
	}
	out, err := h.wishlists.Create(r.Context(), actor, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, out)
}

func (h *Handler) ListWishlists(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	h.list(w, r, actor, r.URL.Query().Get("owner_id"))
}

func (h *Handler) ListUserWishlists(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r)
	h.list(w, r, actor, r.PathValue("userId"))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, actor model.Actor, ownerID string) {
	limit, err := queryInt(r, "limit")
	if err != nil {
		writeError(w, err)
		return
	}
	offset, err := queryInt(r, "offset")
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := h.wishlists.List(r.Context(), actor, port.ListWishlistsInput{OwnerID: ownerID, Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	actor, _ := actorFrom(r)
	out, err := h.wishlists.Get(r.Context(), actor, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) GetSharedWishlist(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r)
	out, err := h.wishlists.GetByKey(r.Context(), actor, r.PathValue("key"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) UpdateWishlist(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in port.UpdateWishlistInput
	if !decode(w, r, &in) {
		return
	}
	out, err := h.wishlists.Update(r.Context(), actor, id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) DeleteWishlist(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.wishlists.Delete(r.Context(), actor, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
