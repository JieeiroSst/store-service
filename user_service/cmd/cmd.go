package cmd

import (
	"os"

	"github.com/JIeeiroSst/user-service/di"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var rootCmd = &cobra.Command{
	Use:   "user-service",
	Short: "user service with gRPC and HTTP gateway",
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(di.Module).Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
