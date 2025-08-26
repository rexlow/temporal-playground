package errors

import (
	"math/rand"
	"time"
)

var Errors = []string{
	"Order not found",
	"Invalid order ID",
	"Order processing failed",
	"Payment provider returns internal error",
	"Customer card provider rejected",
	"Internal server error",
	"Database connection timeout",
	"Unknown error",
}

func GetRandomError() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return Errors[r.Intn(len(Errors))]
}
