package main

import (
	"fmt"
	"time"

	"gopkg.in/robfig/cron.v2"
)

func main() {

	c := cron.New()
	c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	c.AddFunc("TZ=Asia/Bangkok 30 04 * * * *", func() { fmt.Println("Runs at 04:30 Bangkok time every day") })
	c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	c.AddFunc("@every 0h0m1s", func() { fmt.Println("Every second") })
	c.Start()

	// Funcs are invoked in their own goroutine, asynchronously.

	// Funcs may also be added to a running Cron
	c.AddFunc("@daily", func() { fmt.Println("Every day") })

	// Added time to see output
	time.Sleep(10 * time.Second)

	c.Stop() // Stop the scheduler (does not stop any jobs already running).

}
