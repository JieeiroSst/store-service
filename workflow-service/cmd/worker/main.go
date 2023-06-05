package main

import (
	"github.com/JIeeiroSst/workflow-service/internal/activities"
	"github.com/JIeeiroSst/workflow-service/pkg/log"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	err := log.Init("info", "stdout")
	if err != nil {
		panic(err)
	}

	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Error("unable to create Temporal client", err)
	}
	defer c.Close()
	w := worker.New(c, "CART_TASK_QUEUE", worker.Options{})
	a := &activities.Activities{}
	w.RegisterActivity(a.CreateStripeCharge)
	w.RegisterActivity(a.SendAbandonedCartEmail)

	w.RegisterWorkflow(activities.CartWorkflow)
	
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Error("unable to start Worker", err)
	}
}
