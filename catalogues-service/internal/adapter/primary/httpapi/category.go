package httpapi

import (
	"net/http"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type categoryBody struct {
	ParentID        *int64 `json:"parent_id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Description     string `json:"description"`
	Image           string `json:"image"`
	IsPublic        *bool  `json:"is_public"`
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
}

type categoryJSON struct {
	ID                 int64  `json:"id"`
	ParentID           *int64 `json:"parent_id"`
	Name               string `json:"name"`
	Slug               string `json:"slug"`
	Description        string `json:"description"`
	Image              string `json:"image"`
	IsPublic           bool   `json:"is_public"`
	AncestorsArePublic bool   `json:"ancestors_are_public"`
	MetaTitle          string `json:"meta_title"`
	MetaDescription    string `json:"meta_description"`
}

func (b categoryBody) toModel(id int64) model.Category {
	public := true
	if b.IsPublic != nil {
		public = *b.IsPublic
	}
	return model.Category{
		ID: id, ParentID: b.ParentID, Name: b.Name, Slug: b.Slug,
		Description: b.Description, Image: b.Image, IsPublic: public,
		MetaTitle: b.MetaTitle, MetaDescription: b.MetaDescription,
	}
}

func toCategoryJSON(c model.Category) categoryJSON {
	return categoryJSON{
		ID: c.ID, ParentID: c.ParentID, Name: c.Name, Slug: c.Slug,
		Description: c.Description, Image: c.Image, IsPublic: c.IsPublic,
		AncestorsArePublic: c.AncestorsArePublic,
		MetaTitle:          c.MetaTitle, MetaDescription: c.MetaDescription,
	}
}

type CategoryHandler struct{ svc port.CategoryService }

func NewCategoryHandler(svc port.CategoryService) *CategoryHandler { return &CategoryHandler{svc} }

func (h *CategoryHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/categories", h.create)
	mux.HandleFunc("GET /v1/categories", h.list)
	mux.HandleFunc("GET /v1/categories/{id}", h.get)
	mux.HandleFunc("PUT /v1/categories/{id}", h.update)
	mux.HandleFunc("DELETE /v1/categories/{id}", h.delete)
}

func (h *CategoryHandler) create(w http.ResponseWriter, r *http.Request) {
	var b categoryBody
	if !decode(w, r, &b) {
		return
	}
	c, err := h.svc.Create(r.Context(), b.toModel(0))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCategoryJSON(*c))
}

func (h *CategoryHandler) list(w http.ResponseWriter, r *http.Request) {
	cs, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]categoryJSON, len(cs))
	for i, c := range cs {
		out[i] = toCategoryJSON(c)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *CategoryHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCategoryJSON(*c))
}

func (h *CategoryHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b categoryBody
	if !decode(w, r, &b) {
		return
	}
	c, err := h.svc.Update(r.Context(), b.toModel(id))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCategoryJSON(*c))
}

func (h *CategoryHandler) delete(w http.ResponseWriter, r *http.Request) {
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
