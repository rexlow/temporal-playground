package temporal

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerManager struct {
	clientManager *ClientManager
	worker        worker.Worker
}

func NewWorkerManager(options client.Options, taskQueue string) *WorkerManager {
	clientManager := NewClientManager(options)

	if taskQueue == "" {
		log.Panicf("Task queue must be specified")
	}

	w := worker.New(clientManager.GetClient(), taskQueue, worker.Options{
		EnableLoggingInReplay: true,
	})

	return &WorkerManager{
		clientManager: clientManager,
		worker:        w,
	}
}

func (wm *WorkerManager) RegisterWorkflow(workflowFunc any) {
	wm.worker.RegisterWorkflow(workflowFunc)
}

func (wm *WorkerManager) RegisterActivity(activityFunc any) {
	wm.worker.RegisterActivity(activityFunc)
}

func (wm *WorkerManager) Start() error {
	log.Println("Starting Temporal worker...")
	return wm.worker.Run(worker.InterruptCh())
}

func (wm *WorkerManager) Close() {
	wm.clientManager.Close()
}
