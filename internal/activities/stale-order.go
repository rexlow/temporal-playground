package activities

import (
	"context"
	"temporal-playground/internal/models"

	"go.temporal.io/sdk/activity"
)

// FinalizeStaleWorkflow performs final cleanup and resolution
func FinalizeStaleWorkflow(ctx context.Context, request models.StaleWorkflowRequest) error {
	logger := activity.GetLogger(ctx)

	resolution := request.Metadata["resolution"]
	logger.Info("Finalizing stale workflow",
		"originalWorkflowID", request.OriginalWorkflowID,
		"resolution", resolution)

	logger.Info("Stale workflow finalized successfully", "resolution", resolution)

	return nil
}
