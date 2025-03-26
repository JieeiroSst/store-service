package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "account-transaction-service",
	Short: "account-transaction service with gRPC and HTTP gateway",
}

var apiV1Cmd = &cobra.Command{
	Use:   "api-v1",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		runAPIV1()
	},
}

var apiV2Cmd = &cobra.Command{
	Use:   "api-v2",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		runAPIV2()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(apiV1Cmd)
	rootCmd.AddCommand(apiV2Cmd)
}
