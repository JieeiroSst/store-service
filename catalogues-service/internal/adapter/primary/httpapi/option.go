package httpapi

import (
	"net/http"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type optionJSON struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type optionBody struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

func (b optionBody) toModel(id int64) model.Option {
	return model.Option{ID: id, Name: b.Name, Code: b.Code, Type: model.OptionType(b.Type), Required: b.Required}
}

func toOptionJSON(o model.Option) optionJSON {
	return optionJSON{ID: o.ID, Name: o.Name, Code: o.Code, Type: string(o.Type), Required: o.Required}
}

type OptionHandler struct{ svc port.OptionService }

func NewOptionHandler(svc port.OptionService) *OptionHandler { return &OptionHandler{svc} }

func (h *OptionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/options", h.create)
	mux.HandleFunc("GET /v1/options", h.list)
	mux.HandleFunc("GET /v1/options/{id}", h.get)
	mux.HandleFunc("PUT /v1/options/{id}", h.update)
	mux.HandleFunc("DELETE /v1/options/{id}", h.delete)
}

func (h *OptionHandler) create(w http.ResponseWriter, r *http.Request) {
	var b optionBody
	if !decode(w, r, &b) {
		return
	}
	o, err := h.svc.Create(r.Context(), b.toModel(0))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toOptionJSON(*o))
}

func (h *OptionHandler) list(w http.ResponseWriter, r *http.Request) {
	os, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]optionJSON, len(os))
	for i, o := range os {
		out[i] = toOptionJSON(o)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *OptionHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	o, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOptionJSON(*o))
}

func (h *OptionHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b optionBody
	if !decode(w, r, &b) {
		return
	}
	o, err := h.svc.Update(r.Context(), b.toModel(id))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOptionJSON(*o))
}

func (h *OptionHandler) delete(w http.ResponseWriter, r *http.Request) {
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
