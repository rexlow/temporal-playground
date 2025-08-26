package activities

import (
	"context"
	"temporal-playground/internal/models"
	"time"

	"go.temporal.io/sdk/activity"
)

// ConcludeQueryOrder concludes an order with manual resolution
func ConcludeQueryOrder(ctx context.Context, request models.ConcludeQueryOrderRequest) error {
	logger := activity.GetLogger(ctx)

	logger.Info("Concluding query order with manual resolution",
		"orderID", request.OrderID,
		"resolution", request.Resolution,
		"resolvedBy", request.ResolvedBy,
		"resolvedAt", request.ResolvedAt.Format(time.RFC3339))

	return nil
}
