package activities

import (
	"context"
	"math/rand/v2"
	"temporal-playground/internal/errors"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

const (
	FailProbability = 0.8 // 80% failed
)

// heavy operation that will fail randomly
// always retry if failed
func QueryOrder(ctx context.Context, orderID string) error {

	info := activity.GetInfo(ctx)

	// Record heartbeat with detailed progress info
	progressInfo := map[string]any{
		"step":    "initializing",
		"attempt": info.Attempt,
		"orderID": orderID,
	}
	activity.RecordHeartbeat(ctx, progressInfo)

	// Simulate some processing time
	time.Sleep(2 * time.Second)
	probability := rand.Float64()

	// Update progress
	progressInfo["step"] = "processing"
	activity.RecordHeartbeat(ctx, progressInfo)

	time.Sleep(3 * time.Second)

	if probability < FailProbability {
		errorType := errors.GetRandomError()

		errorDetails := map[string]any{
			"errorType":      errorType,
			"attempt":        info.Attempt,
			"orderID":        orderID,
			"activityID":     info.ActivityID,
			"timestamp":      time.Now().Format(time.RFC3339),
			"processingTime": time.Since(info.StartedTime).String(),
		}

		// Record the failure in heartbeat for visibility
		progressInfo["step"] = "failed"
		progressInfo["error"] = errorType
		activity.RecordHeartbeat(ctx, progressInfo)

		return temporal.NewApplicationError(
			errorType,
			"QueryOrderActivity",
			errorDetails,
		)
	}

	progressInfo["step"] = "completed"
	progressInfo["result"] = "success"
	activity.RecordHeartbeat(ctx, progressInfo)

	return nil
}
