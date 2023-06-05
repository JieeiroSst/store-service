package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JIeeiroSst/workflow-service/config"
	"github.com/JIeeiroSst/workflow-service/internal/activities"
	"github.com/JIeeiroSst/workflow-service/pkg/consul"
	"github.com/JIeeiroSst/workflow-service/pkg/log"
	"go.temporal.io/sdk/client"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
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

	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Error("unable to create Temporal client", err)
	}
	defer c.Close()

	workflowID := "CART-" + fmt.Sprintf("%d", time.Now().Unix())

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "CART_TASK_QUEUE",
	}

	state := activities.CartState{Items: make([]activities.CartItem, 0)}
	we, err := c.ExecuteWorkflow(context.Background(), options, activities.CartWorkflow, state)
	if err != nil {
		log.Error("unable to execute workflow", err)
	}

	log.Infof("execute workflow", we.GetRunID())

	update := activities.AddToCartSignal{Route: activities.RouteTypes.ADD_TO_CART, Item: activities.CartItem{ProductId: 0, Quantity: 1}}
	if err = c.SignalWorkflow(context.Background(), workflowID, "", "ADD_TO_CART_CHANNEL", update); err != nil {
		log.Error("signal workflow", err)
	}

	resp, err := c.QueryWorkflow(context.Background(), workflowID, "", "getCart")
	if err != nil {
		log.Error("Unable to query workflow", err)
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		log.Error("Unable to decode query result", err)
	}
	log.Infof("Received query result", "Result", result)
}
