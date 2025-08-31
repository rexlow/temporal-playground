package activities

import (
	"context"

	"go.temporal.io/sdk/activity"
)

func DoSomething(ctx context.Context, input string) (string, error) {
	logger := activity.GetLogger(ctx)

	info := activity.GetInfo(ctx)

	logger.Info("Activity doing something", "input", input)
	logger.Info("Activity info", "info", info)

	return "Result of: " + input, nil
}
