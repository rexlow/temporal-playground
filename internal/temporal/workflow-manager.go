package temporal

import (
	"context"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type WorkflowManager struct {
	clientManager *ClientManager
}

func NewWorkflowManager(options client.Options) *WorkflowManager {
	return &WorkflowManager{
		clientManager: NewClientManager(options),
	}
}

func (wm *WorkflowManager) Close() {
	wm.clientManager.Close()
}

type StartWorkflowOptions struct {
	WorkflowID   string
	TaskQueue    string
	OrderID      string
	Environment  string
	BusinessUnit string
	Priority     string
}

func (wm *WorkflowManager) StartWorkflow(ctx context.Context, options StartWorkflowOptions, workflowFunc any, args ...any) (client.WorkflowRun, error) {

	searchAttributes := map[string]any{
		"CustomStringField":   options.OrderID,
		"CustomKeywordField":  options.Environment,
		"CustomIntField":      1, // use this as priority
		"CustomDatetimeField": time.Now(),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:                    options.WorkflowID,
		TaskQueue:             options.TaskQueue,
		SearchAttributes:      searchAttributes,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY, // allows restart only if previous failed
	}

	return wm.clientManager.GetClient().ExecuteWorkflow(ctx, workflowOptions, workflowFunc, args...)
}

func (wm *WorkflowManager) GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun {
	return wm.clientManager.GetClient().GetWorkflow(ctx, workflowID, runID)
}

func (wm *WorkflowManager) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg any) error {
	return wm.clientManager.GetClient().SignalWorkflow(ctx, workflowID, runID, signalName, arg)
}
