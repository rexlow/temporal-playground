package workflows

import (
	"temporal-playground/internal/activities"
	"temporal-playground/internal/models"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func QueryOrder(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)

	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    3,
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result any
	if err := workflow.ExecuteActivity(ctx, activities.QueryOrder, orderID).Get(ctx, &result); err != nil {

		staleRequest := models.StaleWorkflowRequest{
			OriginalWorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
			OriginalRunID:      workflow.GetInfo(ctx).WorkflowExecution.RunID,
			OrderID:            orderID,
			FailureReason:      "QueryOrderActivity failed after maximum retries",
			FailureTime:        workflow.Now(ctx),
			MaxAttemptsReached: retryPolicy.MaximumAttempts,
			OriginalError:      err.Error(),
			Metadata: map[string]any{
				"workflowType":              "QueryOrderWorkflow",
				"taskQueue":                 "stale-order",
				"namespace":                 workflow.GetInfo(ctx).Namespace,
				"originalWorkflowStartTime": workflow.GetInfo(ctx).WorkflowStartTime,
			},
		}

		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID:        "stale-" + workflow.GetInfo(ctx).WorkflowExecution.ID,
			TaskQueue:         "stale-order",
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON, // Let child workflow continue independently
			SearchAttributes: map[string]any{
				"CustomStringField":   orderID, // Set orderID as CustomStringField
				"CustomKeywordField":  "stale",
				"CustomIntField":      2, // Different priority for stale workflows
				"CustomDatetimeField": workflow.Now(ctx),
			},
		})

		childFuture := workflow.ExecuteChildWorkflow(childCtx, Stale, staleRequest)

		var childExecution workflow.Execution
		if err := childFuture.GetChildWorkflowExecution().Get(ctx, &childExecution); err != nil {
			logger.Error("Failed to start StaleWorkflow", "error", err.Error())
			return err
		}

		return nil
	}

	return nil
}
