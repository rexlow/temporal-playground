package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	hostPort  string
	namespace string
)

var rootCmd = &cobra.Command{
	Use:   "temporal-playground",
	Short: "Temporal workflow application",
	Long:  `A Temporal workflow application with CLI commands to manage workers and workflows.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Temporal namespace (required)")
	rootCmd.PersistentFlags().StringVar(&hostPort, "hostport", "localhost:7233", "Temporal server host:port")
}
