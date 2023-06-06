package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/workflow-service/internal/activities"
	"github.com/gorilla/mux"
	"go.temporal.io/sdk/client"
)

type (
	UpdateEmailRequest struct {
		Email string
	}

	CheckoutRequest struct {
		Email string
	}
)

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]interface{})
	res["products"] = activities.Products

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func CreateCartHandler(w http.ResponseWriter, r *http.Request) {
	workflowID := "CART-" + fmt.Sprintf("%d", time.Now().Unix())

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "CART_TASK_QUEUE",
	}

	cart := activities.CartState{Items: make([]activities.CartItem, 0)}
	we, err := temporal.ExecuteWorkflow(context.Background(), options, activities.CartWorkflow, cart)
	if err != nil {
		WriteError(w, err)
		return
	}

	res := make(map[string]interface{})
	res["cart"] = cart
	res["workflowID"] = we.GetID()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	response, err := temporal.QueryWorkflow(context.Background(), vars["workflowID"], "", "getCart")
	if err != nil {
		WriteError(w, err)
		return
	}
	var res interface{}
	if err := response.Get(&res); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var item activities.CartItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		WriteError(w, err)
		return
	}

	update := activities.AddToCartSignal{Route: activities.RouteTypes.ADD_TO_CART, Item: item}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", activities.SignalChannels.ADD_TO_CART_CHANNEL, update)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var item activities.CartItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		WriteError(w, err)
		return
	}

	update := activities.RemoveFromCartSignal{Route: activities.RouteTypes.REMOVE_FROM_CART, Item: item}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", activities.SignalChannels.REMOVE_FROM_CART_CHANNEL, update)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func UpdateEmailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var body UpdateEmailRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, err)
		return
	}

	updateEmail := activities.UpdateEmailSignal{Route: activities.RouteTypes.UPDATE_EMAIL, Email: body.Email}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", activities.SignalChannels.UPDATE_EMAIL_CHANNEL, updateEmail)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var body CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, err)
		return
	}

	checkout := activities.CheckoutSignal{Route: activities.RouteTypes.CHECKOUT, Email: body.Email}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", activities.SignalChannels.CHECKOUT_CHANNEL, checkout)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["sent"] = true
	json.NewEncoder(w).Encode(res)
}
