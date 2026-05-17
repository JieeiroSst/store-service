package cmd

import (
	"os"

	"github.com/JieeiroSst/authorize-service/internal/infrastructure"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var rootCmd = &cobra.Command{
	Use:   "authorize-service",
	Short: "Authorize service with gRPC and HTTP gateway",
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		app := fx.New(infrastructure.Module)
		app.Run()
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
