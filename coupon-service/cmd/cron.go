package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron"
)

func runCron() {
	log.Print("Starting cron jobs server")

	// Create a new cron scheduler
	c := cron.New()

	// Add cron jobs
	// Basic cron job running every minute
	err := c.AddFunc("* * * * *", func() {
		fmt.Println("Hello, World! Every minute")
	})
	if err != nil {
		log.Fatal("Error adding cron job:", err)
	}

	// Cron job with interval (every 30 seconds)
	err = c.AddFunc("@every 30s", func() {
		fmt.Println("This runs every 30 seconds")
	})
	if err != nil {
		log.Fatal("Error adding cron job:", err)
	}

	// Cron job with specific time (every day at 10:30)
	err = c.AddFunc("30 10 * * *", func() {
		fmt.Println("Daily report time!")
	})
	if err != nil {
		log.Fatal("Error adding cron job:", err)
	}

	// Start the cron scheduler
	c.Start()
	defer c.Stop() // Ensure we clean up on exit

	// Handle shutdown signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down cron jobs...")
}
