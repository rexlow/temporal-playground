package cmd

import (
	"fmt"
	"log"
	"strconv"
	"temporal-playground/internal/temporal"

	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
)

var (
	namespaceName string
	namespaceDesc string
	retentionDays string
)

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Namespace management commands",
	Long:  `Commands to manage Temporal namespaces.`,
}

var registerNamespaceCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new namespace",
	Long:  `Register a new namespace in Temporal server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if namespaceName == "" {
			log.Fatal("Namespace name is required. Use --name flag.")
		}

		retention, err := strconv.Atoi(retentionDays)
		if err != nil {
			log.Fatalf("Invalid retention days: %s. Must be a number.", retentionDays)
		}

		log.Printf("Registering namespace: %s", namespaceName)
		log.Printf("Description: %s", namespaceDesc)
		log.Printf("Retention period: %d days", retention)

		namespaceManager := temporal.NewNamespaceManager(client.Options{
			HostPort:  hostPort,
			Namespace: namespace,
		})
		err = namespaceManager.RegisterNamespace(namespaceName, namespaceDesc, retention)
		if err != nil {
			log.Fatalf("Failed to register namespace: %v", err)
		}

		fmt.Printf("✅ Namespace '%s' registered successfully!\n", namespaceName)
		fmt.Println("You can now use this namespace with:")
		fmt.Printf("  ./temporal-playground worker --namespace %s\n", namespaceName)
		fmt.Printf("  ./temporal-playground client start --namespace %s\n", namespaceName)
	},
}

var listNamespacesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all namespaces",
	Long:  `List all available namespaces in the Temporal server.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Fetching namespaces...")

		namespaceManager := temporal.NewNamespaceManager(client.Options{
			HostPort:  hostPort,
			Namespace: namespace,
		})
		namespaces, err := namespaceManager.ListNamespaces()
		if err != nil {
			log.Fatalf("Failed to list namespaces: %v", err)
		}

		fmt.Println("\n📋 Available Namespaces:")
		fmt.Println("========================")
		for i, ns := range namespaces {
			fmt.Printf("%d. %s\n", i+1, ns)
		}
		fmt.Printf("\nTotal: %d namespaces\n", len(namespaces))
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
	namespaceCmd.AddCommand(registerNamespaceCmd)
	namespaceCmd.AddCommand(listNamespacesCmd)

	// Flags for register command
	registerNamespaceCmd.Flags().StringVar(&namespaceName, "name", "", "Namespace name (required)")
	registerNamespaceCmd.Flags().StringVarP(&namespaceDesc, "description", "d", "", "Namespace description")
	registerNamespaceCmd.Flags().StringVarP(&retentionDays, "retention", "r", "7", "Workflow execution retention period in days")
	registerNamespaceCmd.MarkFlagRequired("name")
}
