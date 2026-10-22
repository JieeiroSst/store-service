package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type stockRecordBody struct {
	PartnerSKU    string  `json:"partner_sku"`
	PriceCurrency string  `json:"price_currency"`
	Price         float64 `json:"price"`
	NumInStock    *int    `json:"num_in_stock"`
}

type stockRecordJSON struct {
	ID            int64     `json:"id"`
	ProductID     int64     `json:"product_id"`
	PartnerSKU    string    `json:"partner_sku"`
	PriceCurrency string    `json:"price_currency"`
	Price         float64   `json:"price"`
	NumInStock    *int      `json:"num_in_stock"`
	NumAllocated  int       `json:"num_allocated"`
	NumAvailable  *int      `json:"num_available"`
	DateCreated   time.Time `json:"date_created"`
	DateUpdated   time.Time `json:"date_updated"`
}

func toStockJSON(s model.StockRecord) stockRecordJSON {
	return stockRecordJSON{
		ID: s.ID, ProductID: s.ProductID, PartnerSKU: s.PartnerSKU, PriceCurrency: s.PriceCurrency,
		Price: s.Price, NumInStock: s.NumInStock, NumAllocated: s.NumAllocated, NumAvailable: s.Available(),
		DateCreated: s.DateCreated, DateUpdated: s.DateUpdated,
	}
}

type StockRecordHandler struct{ svc port.StockRecordService }

func NewStockRecordHandler(svc port.StockRecordService) *StockRecordHandler {
	return &StockRecordHandler{svc}
}

func (h *StockRecordHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/products/{id}/stockrecords", h.list)
	mux.HandleFunc("POST /v1/products/{id}/stockrecords", h.create)
	mux.HandleFunc("GET /v1/stockrecords/{id}", h.get)
	mux.HandleFunc("PUT /v1/stockrecords/{id}", h.update)
	mux.HandleFunc("DELETE /v1/stockrecords/{id}", h.delete)
	mux.HandleFunc("POST /v1/stockrecords/{id}/allocate", h.allocate)
	mux.HandleFunc("POST /v1/stockrecords/{id}/release", h.release)
}

func (h *StockRecordHandler) list(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	rs, err := h.svc.ListByProduct(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]stockRecordJSON, len(rs))
	for i, s := range rs {
		out[i] = toStockJSON(s)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *StockRecordHandler) create(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b stockRecordBody
	if !decode(w, r, &b) {
		return
	}
	s, err := h.svc.Create(r.Context(), model.StockRecord{
		ProductID: id, PartnerSKU: b.PartnerSKU, PriceCurrency: b.PriceCurrency, Price: b.Price, NumInStock: b.NumInStock,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toStockJSON(*s))
}

func (h *StockRecordHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	s, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toStockJSON(*s))
}

func (h *StockRecordHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b stockRecordBody
	if !decode(w, r, &b) {
		return
	}
	s, err := h.svc.Update(r.Context(), model.StockRecord{
		ID: id, PartnerSKU: b.PartnerSKU, PriceCurrency: b.PriceCurrency, Price: b.Price, NumInStock: b.NumInStock,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toStockJSON(*s))
}

func (h *StockRecordHandler) delete(w http.ResponseWriter, r *http.Request) {
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

func (h *StockRecordHandler) allocate(w http.ResponseWriter, r *http.Request) {
	h.adjust(w, r, h.svc.Allocate)
}

func (h *StockRecordHandler) release(w http.ResponseWriter, r *http.Request) {
	h.adjust(w, r, h.svc.Release)
}

func (h *StockRecordHandler) adjust(w http.ResponseWriter, r *http.Request, fn func(context.Context, int64, int) (*model.StockRecord, error)) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var b struct {
		Quantity int `json:"quantity"`
	}
	if !decode(w, r, &b) {
		return
	}
	s, err := fn(r.Context(), id, b.Quantity)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toStockJSON(*s))
}
