package workflows

import (
	"temporal-playground/internal/activities"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func RegisterRecurringPayment(ctx workflow.Context, consentID string) (string, error) {

	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    2,
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
		},
	})

	var result string
	var err error

	// try changing the value of max version here.
	// always increase the max version param instead of lowering it
	version := workflow.GetVersion(ctx, "recurring-payment", workflow.DefaultVersion, 3)
	logger.Info("Using recurring payment version", "version", version)

	switch version {
	case 1:
		err = workflow.ExecuteActivity(ctx, activities.RecurringPaymentV1, consentID).Get(ctx, &result)
	case 2:
		err = workflow.ExecuteActivity(ctx, activities.RecurringPaymentV2, consentID).Get(ctx, &result)
	case 3:
		err = workflow.ExecuteActivity(ctx, activities.RecurringPaymentV3, consentID).Get(ctx, &result)
	}

	return result, err
}
