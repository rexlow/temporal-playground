package activities

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
)

func RecurringPaymentV1(ctx context.Context, consentID string) (string, error) {
	return processPayment(ctx, consentID, 10.00)
}

func RecurringPaymentV2(ctx context.Context, consentID string) (string, error) {
	return processPayment(ctx, consentID, 15.00)
}

func RecurringPaymentV3(ctx context.Context, consentID string) (string, error) {
	return processPayment(ctx, consentID, 20.00)
}

func processPayment(ctx context.Context, consentID string, amount float64) (string, error) {
	logger := activity.GetLogger(ctx)

	time.Sleep(time.Second * 30)

	logger.Info("Processing payment of ", "amount", amount, "consentID", consentID)

	if rand.Float64() < 1 {
		return fmt.Sprintf("Payment of $%.2f processed successfully", amount), nil
	}
	return "", fmt.Errorf("payment processing failed")
}
