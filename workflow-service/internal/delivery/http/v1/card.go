package v1

import (
	"encoding/json"
	"net/http"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/activities"
	"github.com/go-chi/chi/v5"
)

func (h *Http) GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	// res := make(map[string]interface{})
	// res["products"] = activities.Products

	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(res)
}

func (h *Http) CreateCartHandler(w http.ResponseWriter, r *http.Request) {
	res, err := h.Usecase.CreateCard()
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *Http) GetCartHandler(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")

	res, err := h.Usecase.Cards.GetCard(workflowID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *Http) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	workflowID := chi.URLParam(r, "workflowID")
	var item activities.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		WriteError(w, err)
		return
	}

	if err := h.Usecase.Cards.AddToCard(workflowID, item); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func (h *Http) RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	var item activities.CartItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := h.Usecase.Cards.RemoveFromCard(workflowID, item); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func (h *Http) UpdateEmailHandler(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")

	var body dto.UpdateEmailRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := h.Usecase.Cards.UpdateEmailToCard(workflowID, body); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func (h *Http) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	var body dto.CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := h.Usecase.Cards.CheckoutCard(workflowID, body); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["sent"] = true
	json.NewEncoder(w).Encode(res)
}
