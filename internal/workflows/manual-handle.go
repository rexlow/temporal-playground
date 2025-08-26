package workflows

import (
	"temporal-playground/internal/activities"
	"temporal-playground/internal/models"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ManualHandleOrderWorkflow handles orders that require manual intervention
// This workflow waits indefinitely until a manual signal is received
func ManualHandleOrder(ctx workflow.Context, request models.ManualHandleRequest) error {

	var (
		logger         = workflow.GetLogger(ctx)
		selector       = workflow.NewSelector(ctx)
		resolveChannel = workflow.GetSignalChannel(ctx, "resolve-manual-order")
		resolveSignal  string
	)

	// Wait indefinitely for manual resolution signal
	selector.AddReceive(resolveChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &resolveSignal)
	})
	selector.Select(ctx)

	if resolveSignal != "" {

		activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 5 * time.Minute,
			HeartbeatTimeout:    30 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second * 5,
				BackoffCoefficient: 1.5,
				MaximumInterval:    time.Minute * 2,
				MaximumAttempts:    3,
			},
		})
		if err := workflow.ExecuteActivity(activityCtx, activities.ConcludeQueryOrder, models.ConcludeQueryOrderRequest{
			OrderID:            request.OrderID,
			Resolution:         resolveSignal,
			OriginalWorkflowID: request.OriginalWorkflowID,
			ResolvedAt:         workflow.Now(ctx),
			ResolvedBy:         "manual-intervention",
		}).Get(ctx, nil); err != nil {
			logger.Error("Failed to conclude order", "error", err.Error())
			return err
		}

		logger.Info("ManualHandle workflow completed successfully",
			"orderID", request.OrderID,
			"resolution", resolveSignal)
	}

	return nil
}
