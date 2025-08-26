package models

import "time"

// Resolution status constants
const (
	ResolutionSuccess       = "success"
	ResolutionFailed        = "failed"
	ResolutionRetrySuccess  = "retry-succeeded"
	ResolutionManualResolve = "manual-resolve"
	ResolutionMovedToStale  = "moved-to-stale"
	ResolutionMovedToManual = "moved-to-manual"
)

// StaleWorkflowRequest represents the data for a workflow that has failed after all retries
type StaleWorkflowRequest struct {
	OriginalWorkflowID string         `json:"originalWorkflowID"`
	OriginalRunID      string         `json:"originalRunID"`
	OrderID            string         `json:"orderID"`
	FailureReason      string         `json:"failureReason"`
	FailureTime        time.Time      `json:"failureTime"`
	MaxAttemptsReached int32          `json:"maxAttemptsReached"`
	OriginalError      string         `json:"originalError,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// ManualHandleRequest represents the data for a workflow that needs manual intervention
type ManualHandleRequest struct {
	OriginalWorkflowID string         `json:"originalWorkflowID"`
	OriginalRunID      string         `json:"originalRunID"`
	OrderID            string         `json:"orderID"`
	FailureReason      string         `json:"failureReason"`
	StaleWorkflowID    string         `json:"staleWorkflowID"`
	StaleFailureTime   time.Time      `json:"staleFailureTime"`
	OriginalError      string         `json:"originalError,omitempty"`
	StaleRetryError    string         `json:"staleRetryError,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// ConcludeQueryOrderRequest represents the data for concluding an order
type ConcludeQueryOrderRequest struct {
	OrderID            string    `json:"orderID"`
	Resolution         string    `json:"resolution"`
	OriginalWorkflowID string    `json:"originalWorkflowID"`
	ResolvedAt         time.Time `json:"resolvedAt"`
	ResolvedBy         string    `json:"resolvedBy,omitempty"`
}
