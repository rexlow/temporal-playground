package cmd

import (
	"log"
	"sync"
	"temporal-playground/internal/activities"
	"temporal-playground/internal/temporal"
	"temporal-playground/internal/workflows"

	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
)

// start all workers: testing purpose only, do not do this in prod
// we should run a single worker per container
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start all Temporal workers",
	Long:  `Start all Temporal workers to process workflows and activities.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Starting Temporal worker in namespace: %s", namespace)

		// Create all workers
		workers := []*temporal.WorkerManager{
			temporal.NewWorkerManager(client.Options{
				HostPort:  hostPort,
				Namespace: namespace,
			}, "query-order"),
			temporal.NewWorkerManager(client.Options{
				HostPort:  hostPort,
				Namespace: namespace,
			}, "stale-order"),
			temporal.NewWorkerManager(client.Options{
				HostPort:  hostPort,
				Namespace: namespace,
			}, "manual-handle"),
		}

		// Ensure all workers are closed on exit
		defer func() {
			for _, worker := range workers {
				worker.Close()
			}
		}()

		// Register workflows and activities for each worker
		// Query Order Worker (index 0)
		workers[0].RegisterWorkflow(workflows.QueryOrder)
		workers[0].RegisterActivity(activities.QueryOrder)
		workers[0].RegisterActivity(activities.FinalizeStaleWorkflow)
		workers[0].RegisterActivity(activities.ConcludeQueryOrder)

		// Stale Order Worker (index 1)
		workers[1].RegisterWorkflow(workflows.Stale)
		workers[1].RegisterActivity(activities.QueryOrder)
		workers[1].RegisterActivity(activities.FinalizeStaleWorkflow)
		workers[1].RegisterActivity(activities.ConcludeQueryOrder)

		// Manual Handle Worker (index 2)
		workers[2].RegisterWorkflow(workflows.ManualHandleOrder)
		workers[2].RegisterActivity(activities.QueryOrder)
		workers[2].RegisterActivity(activities.FinalizeStaleWorkflow)
		workers[2].RegisterActivity(activities.ConcludeQueryOrder)

		var wg sync.WaitGroup
		workerNames := []string{"query-order", "stale-order", "manual-handle"}

		for i, worker := range workers {
			wg.Add(1)
			go func(w *temporal.WorkerManager, name string) {
				defer wg.Done()
				log.Printf("Starting %s worker...", name)
				if err := w.Start(); err != nil {
					log.Fatalf("Unable to start %s worker: %v", name, err)
				}
				log.Printf("%s worker started successfully", name)
			}(worker, workerNames[i])
		}

		log.Println("All workers started successfully. Press Ctrl+C to stop.")
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
