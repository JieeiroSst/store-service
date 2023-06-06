package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JIeeiroSst/workflow-service/config"
	"github.com/JIeeiroSst/workflow-service/internal/activities"
	"github.com/JIeeiroSst/workflow-service/pkg/consul"
	"github.com/JIeeiroSst/workflow-service/pkg/log"
	"github.com/bojanz/httpx"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.temporal.io/sdk/client"
)

type (
	ErrorResponse struct {
		Message string
	}

	UpdateEmailRequest struct {
		Email string
	}

	CheckoutRequest struct {
		Email string
	}
)

var (
	HTTPPort = os.Getenv("PORT")
	temporal client.Client
	conf     *config.Config
	dirEnv   *config.Dir
	err      error
)

func main() {
	err := log.Init("info", "stdout")
	if err != nil {
		panic(err)
	}

	nodeEnv := os.Getenv("NODE_ENV")

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Error("", err)
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Error("", err)
		}
	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Error("", err)
		}
	}

	temporal, err = client.NewClient(client.Options{})
	if err != nil {
		log.Error("unable to create Temporal client", err)
	}
	log.Info("Temporal client connected")

	r := mux.NewRouter()
	r.Handle("/products", http.HandlerFunc(GetProductsHandler)).Methods("GET")
	r.Handle("/cart", http.HandlerFunc(CreateCartHandler)).Methods("POST")
	r.Handle("/cart/{workflowID}", http.HandlerFunc(GetCartHandler)).Methods("GET")
	r.Handle("/cart/{workflowID}/add", http.HandlerFunc(AddToCartHandler)).Methods("PUT")
	r.Handle("/cart/{workflowID}/remove", http.HandlerFunc(RemoveFromCartHandler)).Methods("PUT")
	r.Handle("/cart/{workflowID}/checkout", http.HandlerFunc(CheckoutHandler)).Methods("PUT")
	r.Handle("/cart/{workflowID}/email", http.HandlerFunc(UpdateEmailHandler)).Methods("PUT")

	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	var cors = handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))

	http.Handle("/", cors(r))
	server := httpx.NewServer(":"+HTTPPort, http.DefaultServeMux)
	server.WriteTimeout = time.Second * 240

	log.Infof("Starting server on port: " + HTTPPort)

	err = server.Start()
	if err != nil {
		log.Error("", err)
	}
}

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]interface{})
	// res["products"] = activities.Products

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

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	res := ErrorResponse{Message: "Endpoint not found"}
	json.NewEncoder(w).Encode(res)
}

func WriteError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	res := ErrorResponse{Message: err.Error()}
	json.NewEncoder(w).Encode(res)
}
