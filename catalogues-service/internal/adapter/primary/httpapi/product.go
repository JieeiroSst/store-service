package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type productBody struct {
	Structure       string   `json:"structure"`
	UPC             *string  `json:"upc"`
	Title           string   `json:"title"`
	Slug            string   `json:"slug"`
	Description     string   `json:"description"`
	Rating          *float64 `json:"rating"`
	IsDiscountable  *bool    `json:"is_discountable"`
	IsPublic        *bool    `json:"is_public"`
	ParentID        *int64   `json:"parent_id"`
	ProductClassID  *int64   `json:"product_class_id"`
	MetaTitle       string   `json:"meta_title"`
	MetaDescription string   `json:"meta_description"`
	CategoryIDs     []int64  `json:"category_ids"`
	OptionIDs       []int64  `json:"option_ids"`
}

func orTrue(b *bool) bool { return b == nil || *b }

func (b productBody) toModel(id int64) model.Product {
	return model.Product{
		ID: id, Structure: model.Structure(b.Structure), UPC: b.UPC, Title: b.Title,
		Slug: b.Slug, Description: b.Description, Rating: b.Rating,
		IsDiscountable: orTrue(b.IsDiscountable), IsPublic: orTrue(b.IsPublic),
		ParentID: b.ParentID, ProductClassID: b.ProductClassID,
		MetaTitle: b.MetaTitle, MetaDescription: b.MetaDescription,
		CategoryIDs: b.CategoryIDs, OptionIDs: b.OptionIDs,
	}
}

type imageJSON struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	Original     string    `json:"original"`
	Caption      string    `json:"caption"`
	DisplayOrder int       `json:"display_order"`
	DateCreated  time.Time `json:"date_created"`
}

type recommendationJSON struct {
	ProductID int64 `json:"product_id"`
	Ranking   int   `json:"ranking"`
}

type productJSON struct {
	ID              int64                `json:"id"`
	Structure       string               `json:"structure"`
	UPC             *string              `json:"upc"`
	Title           string               `json:"title"`
	Slug            string               `json:"slug"`
	Description     string               `json:"description"`
	Rating          *float64             `json:"rating"`
	IsDiscountable  bool                 `json:"is_discountable"`
	IsPublic        bool                 `json:"is_public"`
	ParentID        *int64               `json:"parent_id"`
	ProductClassID  *int64               `json:"product_class_id"`
	MetaTitle       string               `json:"meta_title"`
	MetaDescription string               `json:"meta_description"`
	DateCreated     time.Time            `json:"date_created"`
	DateUpdated     time.Time            `json:"date_updated"`
	CategoryIDs     []int64              `json:"category_ids"`
	OptionIDs       []int64              `json:"option_ids"`
	Images          []imageJSON          `json:"images"`
	Recommendations []recommendationJSON `json:"recommendations"`
}

func toImageJSON(m model.ProductImage) imageJSON {
	return imageJSON{
		ID: m.ID, ProductID: m.ProductID, Original: m.Original, Caption: m.Caption,
		DisplayOrder: m.DisplayOrder, DateCreated: m.DateCreated,
	}
}

func toProductJSON(p model.Product) productJSON {
	out := productJSON{
		ID: p.ID, Structure: string(p.Structure), UPC: p.UPC, Title: p.Title, Slug: p.Slug,
		Description: p.Description, Rating: p.Rating, IsDiscountable: p.IsDiscountable,
		IsPublic: p.IsPublic, ParentID: p.ParentID, ProductClassID: p.ProductClassID,
		MetaTitle: p.MetaTitle, MetaDescription: p.MetaDescription,
		DateCreated: p.DateCreated, DateUpdated: p.DateUpdated,
		CategoryIDs: p.CategoryIDs, OptionIDs: p.OptionIDs,
		Images:          make([]imageJSON, len(p.Images)),
		Recommendations: make([]recommendationJSON, len(p.Recommendations)),
	}
	if out.CategoryIDs == nil {
		out.CategoryIDs = []int64{}
	}
	if out.OptionIDs == nil {
		out.OptionIDs = []int64{}
	}
	for i, m := range p.Images {
		out.Images[i] = toImageJSON(m)
	}
	for i, r := range p.Recommendations {
		out.Recommendations[i] = recommendationJSON{ProductID: r.RecommendedID, Ranking: r.Ranking}
	}
	return out
}

type ProductHandler struct{ svc port.ProductService }

func NewProductHandler(svc port.ProductService) *ProductHandler { return &ProductHandler{svc} }

func (h *ProductHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/products", h.create)
	mux.HandleFunc("GET /v1/products", h.list)
	mux.HandleFunc("GET /v1/products/{id}", h.get)
	mux.HandleFunc("PUT /v1/products/{id}", h.update)
	mux.HandleFunc("DELETE /v1/products/{id}", h.delete)
	mux.HandleFunc("POST /v1/products/{id}/images", h.addImage)
	mux.HandleFunc("DELETE /v1/products/{id}/images/{image_id}", h.deleteImage)
	mux.HandleFunc("PUT /v1/products/{id}/recommendations", h.setRecommendations)
}

func (h *ProductHandler) create(w http.ResponseWriter, r *http.Request) {
	var b productBody
	if !decode(w, r, &b) {
		return
	}
	p, err := h.svc.Create(r.Context(), b.toModel(0))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toProductJSON(*p))
}

func (h *ProductHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := model.ProductFilter{Query: q.Get("q"), OnlyPublic: q.Get("public") == "true"}
	if v := q.Get("category_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, model.Invalid("category_id must be a positive integer"))
			return
		}
		f.CategoryID = &id
	}
	for name, dst := range map[string]*int{"limit": &f.Limit, "offset": &f.Offset} {
		if v := q.Get(name); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				writeError(w, model.Invalid("%s must be a non-negative integer", name))
				return
			}
			*dst = n
		}
	}
	ps, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]productJSON, len(ps))
	for i, p := range ps {
		out[i] = toProductJSON(p)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProductHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	p, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductJSON(*p))
}

func (h *ProductHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b productBody
	if !decode(w, r, &b) {
		return
	}
	p, err := h.svc.Update(r.Context(), b.toModel(id))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductJSON(*p))
}

func (h *ProductHandler) delete(w http.ResponseWriter, r *http.Request) {
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

func (h *ProductHandler) addImage(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b struct {
		Original     string `json:"original"`
		Caption      string `json:"caption"`
		DisplayOrder int    `json:"display_order"`
	}
	if !decode(w, r, &b) {
		return
	}
	img, err := h.svc.AddImage(r.Context(), model.ProductImage{
		ProductID: id, Original: b.Original, Caption: b.Caption, DisplayOrder: b.DisplayOrder,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toImageJSON(*img))
}

func (h *ProductHandler) deleteImage(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	imageID, ok := pathID(w, r, "image_id")
	if !ok {
		return
	}
	if err := h.svc.DeleteImage(r.Context(), id, imageID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) setRecommendations(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b struct {
		Recommendations []recommendationJSON `json:"recommendations"`
	}
	if !decode(w, r, &b) {
		return
	}
	recs := make([]model.Recommendation, len(b.Recommendations))
	for i, rc := range b.Recommendations {
		recs[i] = model.Recommendation{RecommendedID: rc.ProductID, Ranking: rc.Ranking}
	}
	p, err := h.svc.SetRecommendations(r.Context(), id, recs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductJSON(*p))
}
