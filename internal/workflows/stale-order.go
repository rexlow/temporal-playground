package workflows

import (
	"temporal-playground/internal/activities"
	"temporal-playground/internal/models"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	MaximumAttempts      = 3 // Limited retries for stale workflow activities
	RetryQueryOrderCount = 1 // Number of retries when retrying the original job
	RetryDuration        = 1 * time.Minute
)

// StaleWorkflow handles workflows that have failed after all retries
// This workflow keeps track of failed workflows and can be used for manual intervention,
// reporting, or implementing custom retry logic outside of the original workflow
func Stale(ctx workflow.Context, request models.StaleWorkflowRequest) error {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 5,
			BackoffCoefficient: 1.5,
			MaximumInterval:    time.Minute * 2,
			MaximumAttempts:    MaximumAttempts,
		},
	})

	var (
		logger         = workflow.GetLogger(ctx)
		retryTimer     = workflow.NewTimer(ctx, RetryDuration)
		selector       = workflow.NewSelector(ctx)
		resolveChannel = workflow.GetSignalChannel(ctx, "resolve-stale-workflow")
		resolveSignal  string
		timerFired     bool
	)

	selector.AddReceive(resolveChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &resolveSignal)
		logger.Info("Received resolve signal for stale workflow", "signal", resolveSignal)
	})

	selector.AddFuture(retryTimer, func(f workflow.Future) {
		timerFired = true
		logger.Info("Retry timer expired - attempting to retry original job", "originalWorkflowID", request.OriginalWorkflowID)
	})
	selector.Select(ctx)

	// Handle the result
	if resolveSignal != "" {
		logger.Info("Manually resolved stale workflow", "resolution", resolveSignal)
	} else if timerFired {

		retryCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 2 * time.Minute,
			HeartbeatTimeout:    10 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second * 30,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute * 5,
				MaximumAttempts:    RetryQueryOrderCount, // Give it a few more tries
			},
		})

		var retryResult any
		if err := workflow.ExecuteActivity(retryCtx, activities.QueryOrder, request.OrderID).Get(ctx, &retryResult); err != nil {

			manualRequest := models.ManualHandleRequest{
				OriginalWorkflowID: request.OriginalWorkflowID,
				OriginalRunID:      request.OriginalRunID,
				OrderID:            request.OrderID,
				FailureReason:      "Stale workflow retry failed after delay",
				StaleFailureTime:   workflow.Now(ctx),
				OriginalError:      request.OriginalError,
				StaleRetryError:    err.Error(),
				Metadata: map[string]any{
					"workflowType":           "StaleWorkflow",
					"staleWorkflowStartTime": request.FailureTime,
					"originalFailureReason":  request.FailureReason,
					"escalationLevel":        "manual-intervention-required",
				},
			}

			// Start ManualHandle workflow as child workflow
			childWorkflowOptions := workflow.ChildWorkflowOptions{
				WorkflowID:        "manual-" + workflow.GetInfo(ctx).WorkflowExecution.ID,
				TaskQueue:         "manual-handle",
				ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
				SearchAttributes: map[string]any{
					"CustomStringField":   request.OrderID,
					"CustomKeywordField":  "manual",
					"CustomIntField":      3, // Highest priority for manual workflows
					"CustomDatetimeField": workflow.Now(ctx),
				},
			}
			childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)
			_ = workflow.ExecuteChildWorkflow(childCtx, ManualHandleOrder, manualRequest)

			resolveSignal = "moved-to-manual-handle"
		} else {
			resolveSignal = "retry-succeeded"
		}
	}

	// Final step: Mark as resolved
	finalRequest := request
	finalRequest.Metadata["resolvedAt"] = time.Now()
	finalRequest.Metadata["resolution"] = resolveSignal
	if resolveSignal == "" {
		finalRequest.Metadata["resolution"] = "unknown"
	}

	var finalResult any
	if err := workflow.ExecuteActivity(ctx, activities.FinalizeStaleWorkflow, finalRequest).Get(ctx, &finalResult); err != nil {
		logger.Error("Failed to finalize stale workflow", "error", err.Error())
		return err
	}

	return nil
}
