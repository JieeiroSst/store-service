package httpapi

import (
	"net/http"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type productClassBody struct {
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	RequiresShipping bool    `json:"requires_shipping"`
	TrackStock       bool    `json:"track_stock"`
	OptionIDs        []int64 `json:"option_ids"`
}

type productClassJSON struct {
	ID int64 `json:"id"`
	productClassBody
}

func (b productClassBody) toModel(id int64) model.ProductClass {
	return model.ProductClass{
		ID: id, Name: b.Name, Slug: b.Slug,
		RequiresShipping: b.RequiresShipping, TrackStock: b.TrackStock, OptionIDs: b.OptionIDs,
	}
}

func toProductClassJSON(c model.ProductClass) productClassJSON {
	ids := c.OptionIDs
	if ids == nil {
		ids = []int64{}
	}
	return productClassJSON{ID: c.ID, productClassBody: productClassBody{
		Name: c.Name, Slug: c.Slug, RequiresShipping: c.RequiresShipping,
		TrackStock: c.TrackStock, OptionIDs: ids,
	}}
}

type ProductClassHandler struct{ svc port.ProductClassService }

func NewProductClassHandler(svc port.ProductClassService) *ProductClassHandler {
	return &ProductClassHandler{svc}
}

func (h *ProductClassHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/product-classes", h.create)
	mux.HandleFunc("GET /v1/product-classes", h.list)
	mux.HandleFunc("GET /v1/product-classes/{id}", h.get)
	mux.HandleFunc("PUT /v1/product-classes/{id}", h.update)
	mux.HandleFunc("DELETE /v1/product-classes/{id}", h.delete)
}

func (h *ProductClassHandler) create(w http.ResponseWriter, r *http.Request) {
	var b productClassBody
	if !decode(w, r, &b) {
		return
	}
	c, err := h.svc.Create(r.Context(), b.toModel(0))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toProductClassJSON(*c))
}

func (h *ProductClassHandler) list(w http.ResponseWriter, r *http.Request) {
	cs, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]productClassJSON, len(cs))
	for i, c := range cs {
		out[i] = toProductClassJSON(c)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProductClassHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductClassJSON(*c))
}

func (h *ProductClassHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b productClassBody
	if !decode(w, r, &b) {
		return
	}
	c, err := h.svc.Update(r.Context(), b.toModel(id))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductClassJSON(*c))
}

func (h *ProductClassHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
