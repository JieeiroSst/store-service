package main

import (
	"os"
	"strings"

	"github.com/JIeeiroSst/workflow-service/config"
	"github.com/JIeeiroSst/workflow-service/internal/activities"
	"github.com/JIeeiroSst/workflow-service/pkg/consul"
	"github.com/JIeeiroSst/workflow-service/pkg/log"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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
