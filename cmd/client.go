package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"temporal-playground/internal/temporal"
	"temporal-playground/internal/workflows"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
)

const (
	SimulatePaymentCount = 500 // per second
)

var (
	workflowID            string
	orderID               string
	environment           string
	businessUnit          string
	priority              string
	recurringPaymentTerms int
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Temporal client commands",
	Long:  `Commands to interact with Temporal workflows as a client.`,
}

var startWorkflowCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a workflow execution",
	Long:  `Start a new workflow execution.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Starting workflow with ID: %s in namespace: %s", workflowID, namespace)

		workflowManager := temporal.NewWorkflowManager(client.Options{
			HostPort:  hostPort,
			Namespace: namespace,
		})
		defer workflowManager.Close()

		orderIDFlag, _ := cmd.Flags().GetString("order-id")
		if orderIDFlag == "" {
			orderIDFlag = uuid.NewString()
		}

		workflowID = fmt.Sprintf("payment-%s", orderIDFlag)

		workflowOptions := temporal.StartWorkflowOptions{
			WorkflowID:   workflowID,
			TaskQueue:    QueueQueryOrder,
			OrderID:      orderIDFlag,
			Environment:  environment,
			BusinessUnit: businessUnit,
			Priority:     priority,
		}

		_, err := workflowManager.StartWorkflow(
			context.Background(),
			workflowOptions,
			workflows.QueryOrder,
			orderIDFlag,
		)
		if err != nil {
			// Check if it's a duplicate workflow error
			if strings.Contains(err.Error(), "WorkflowExecutionAlreadyStarted") {
				log.Printf("Order %s is already being processed (workflow %s) - cannot start duplicate", orderIDFlag, workflowID)
				return
			}
			log.Fatalf("Unable to execute workflow: %v", err)
		}
	},
}

var simulatePaymentWorkflowCmd = &cobra.Command{
	Use:   "simulate-payment",
	Short: "Simulate payment activities",
	Long:  `Simulate long running payment activities that creates orders every second until interrupted - press Ctrl+C to stop`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowManager := temporal.NewWorkflowManager(client.Options{
			HostPort:  hostPort,
			Namespace: namespace,
		})
		defer workflowManager.Close()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		totalWorkflows := 0

		for {
			select {
			case <-ticker.C:

				transactionCount := rand.Intn(SimulatePaymentCount) + 1

				log.Printf("Processing batch %d: received %d transactions", totalWorkflows+1, transactionCount)

				for range transactionCount {
					orderID := uuid.New().String()

					// Use orderID as the workflow ID to prevent duplicates
					// This ensures only one workflow per order can run at a time
					workflowID := fmt.Sprintf("payment-%s", orderID)

					workflowOptions := temporal.StartWorkflowOptions{
						WorkflowID:   workflowID,
						TaskQueue:    QueueQueryOrder,
						OrderID:      orderID,
						Environment:  environment,
						BusinessUnit: businessUnit,
						Priority:     priority,
					}

					_, err := workflowManager.StartWorkflow(
						context.Background(),
						workflowOptions,
						workflows.QueryOrder,
						orderID,
					)
					if err != nil {
						// Check if it's a duplicate workflow error
						if strings.Contains(err.Error(), "WorkflowExecutionAlreadyStarted") {
							log.Printf("Order %s is already being processed (workflow %s) - skipping", orderID, workflowID)
							continue
						}
						log.Printf("Failed to start workflow %s: %v", workflowID, err)
						continue
					}
				}

				totalWorkflows += transactionCount
				log.Printf("Total workflows started: %d", totalWorkflows)

			case <-interrupt:
				log.Printf("\nReceived interrupt signal. Stopping simulation...")
				log.Printf("Total workflows started: %d", totalWorkflows)
				return
			}
		}
	},
}

var signalManualWorkflowCmd = &cobra.Command{
	Use:   "signal-manual",
	Short: "Send a signal to resolve a manual workflow",
	Long:  `Send a resolution signal to a running manual workflow.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Sending signal to manual workflow with ID: %s in namespace: %s", workflowID, namespace)

		workflowManager := temporal.NewWorkflowManager(client.Options{
			HostPort:  hostPort,
			Namespace: namespace,
		})
		defer workflowManager.Close()

		// Get resolution from command line or default
		resolution := "manual-resolution"
		if len(args) > 0 {
			resolution = args[0]
		}

		err := workflowManager.SignalWorkflow(
			context.Background(),
			workflowID,
			"",
			"resolve-manual-order",
			resolution,
		)
		if err != nil {
			log.Fatalf("Unable to signal workflow: %v", err)
		}

		log.Printf("Successfully sent signal '%s' to manual workflow %s", resolution, workflowID)
	},
}

var createRecurringPaymentCmd = &cobra.Command{
	Use:   "create-recurring-payment",
	Short: "Create a recurring payment for a customer",
	Run: func(cmd *cobra.Command, args []string) {

		orderIDFlag, _ := cmd.Flags().GetString("order-id")
		if orderIDFlag == "" {
			orderIDFlag = uuid.NewString()
		}

		workflowManager := temporal.NewWorkflowManager(client.Options{
			HostPort:  hostPort,
			Namespace: namespace,
		})
		defer workflowManager.Close()

		scheduleHandle, err := workflowManager.StartScheduledWorkflow(context.Background(), temporal.ScheduleWorkflowOptions{
			RemainingActions: recurringPaymentTerms,
			Specs: client.ScheduleSpec{
				TimeZoneName: "Asia/Kuala_Lumpur",
				Intervals: []client.ScheduleIntervalSpec{
					{
						Every: 1 * time.Minute,
					},
				},
			},
			StartWorkflowOptions: temporal.StartWorkflowOptions{
				WorkflowID:   orderIDFlag,
				TaskQueue:    QueueRecurringSchedule,
				OrderID:      orderIDFlag,
				Environment:  environment,
				BusinessUnit: businessUnit,
				Priority:     priority,
			},
		}, workflows.RegisterRecurringPayment, orderIDFlag)
		if err != nil {
			log.Fatalf("Failed to schedule recurring payment workflow: %v", err)
		}

		log.Printf("Successfully scheduled recurring payment workflow: %s (schedule ID: %s)", orderIDFlag, scheduleHandle.GetID())
	},
}

var cancelRecurringPaymentCmd = &cobra.Command{
	Use:   "cancel-recurring-payment",
	Short: "Cancel a recurring payment workflow",
	Run: func(cmd *cobra.Command, args []string) {

		ctx := context.Background()

		workflowManager := temporal.NewWorkflowManager(client.Options{
			HostPort:  hostPort,
			Namespace: namespace,
		})
		defer workflowManager.Close()

		if orderID == "" {
			log.Fatalf("Order ID is required")
		}

		scheduleHandle := workflowManager.GetScheduleHandle(ctx, orderID)
		if scheduleHandle == nil {
			log.Fatalf("Unable to cancel workflow")
		}

		if err := scheduleHandle.Delete(ctx); err != nil {
			log.Fatalf("Unable to cancel workflow: %v", err)
		}

		log.Printf("Successfully cancelled recurring payment workflow %s", workflowID)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.AddCommand(startWorkflowCmd)
	clientCmd.AddCommand(simulatePaymentWorkflowCmd)
	clientCmd.AddCommand(signalManualWorkflowCmd)
	clientCmd.AddCommand(createRecurringPaymentCmd)
	clientCmd.AddCommand(cancelRecurringPaymentCmd)

	// Flags for start-workflow command
	startWorkflowCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "query-order-workflow-default", "Workflow ID")
	startWorkflowCmd.Flags().StringVarP(&orderID, "order-id", "o", "", "Order ID to process")
	startWorkflowCmd.Flags().StringVarP(&environment, "environment", "e", "development", "Environment (dev/staging/prod)")
	startWorkflowCmd.Flags().StringVarP(&businessUnit, "business-unit", "b", "retail", "Business unit")
	startWorkflowCmd.Flags().StringVarP(&priority, "priority", "p", "normal", "Priority level (low/normal/high/urgent)")

	// Flags for start-workflow command
	simulatePaymentWorkflowCmd.Flags().StringVarP(&environment, "environment", "e", "development", "Environment (dev/staging/prod)")
	simulatePaymentWorkflowCmd.Flags().StringVarP(&businessUnit, "business-unit", "b", "retail", "Business unit")
	simulatePaymentWorkflowCmd.Flags().StringVarP(&priority, "priority", "p", "normal", "Priority level (low/normal/high/urgent)")

	// Flags for signal-manual-workflow command
	signalManualWorkflowCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "Manual workflow ID to signal")
	signalManualWorkflowCmd.MarkFlagRequired("workflow-id")

	createRecurringPaymentCmd.Flags().IntVarP(&recurringPaymentTerms, "terms", "r", 0, "Number of payment terms (0 means infinite)")
	createRecurringPaymentCmd.Flags().StringVarP(&orderID, "order-id", "o", "", "Use this as consent ID")
	createRecurringPaymentCmd.Flags().StringVarP(&environment, "environment", "e", "development", "Environment (dev/staging/prod)")
	createRecurringPaymentCmd.Flags().StringVarP(&businessUnit, "business-unit", "b", "retail", "Business unit")
	createRecurringPaymentCmd.Flags().StringVarP(&priority, "priority", "p", "normal", "Priority level (low/normal/high/urgent)")

	cancelRecurringPaymentCmd.Flags().StringVarP(&orderID, "order-id", "o", "", "Order ID to cancel")
	cancelRecurringPaymentCmd.MarkFlagRequired("order-id")

}
