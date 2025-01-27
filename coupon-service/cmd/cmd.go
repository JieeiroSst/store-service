package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "coupon-service",
	Short: "Coupon service with gRPC and HTTP gateway and consumer queue and cron jobs and subscriber",
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		runAPI()
	},
}

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Start the cron jobs server",
	Run: func(cmd *cobra.Command, args []string) {
		runCron()
	},
}

var consumerCmd = &cobra.Command{
	Use:   "consumer",
	Short: "Start the consumer server",
	Run: func(cmd *cobra.Command, args []string) {
		runConsumer()
	},
}

var subscriberCmd = &cobra.Command{
	Use:   "subscriber",
	Short: "Start the subscriber server",
	Run: func(cmd *cobra.Command, args []string) {
		runSubscriber()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(cronCmd)
	rootCmd.AddCommand(consumerCmd)
	rootCmd.AddCommand(subscriberCmd)
}
