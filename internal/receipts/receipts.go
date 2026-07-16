// Package receipts stores a durable local mirror of FortSignal decisions.
package receipts

import (
	"context"
	"time"
)

// Record is one governance outcome bound to a vault path when applicable.
type Record struct {
	ID           string    `json:"id"`
	SignalID     string    `json:"signalId,omitempty"`
	Decision     string    `json:"decision"`
	Reason       string    `json:"reason,omitempty"`
	Action       string    `json:"action"`
	Path         string    `json:"path,omitempty"`
	ContentHash  string    `json:"contentHash,omitempty"`
	AgentID      string    `json:"agentId,omitempty"`
	DelegationID string    `json:"delegationId,omitempty"`
	PolicyID     string    `json:"policyId,omitempty"`
	VerifiedBy   string    `json:"verifiedBy,omitempty"`
	VerifiedAt   time.Time `json:"verifiedAt"`
	VaultID      string    `json:"vaultId,omitempty"`
}

// Store is append-mostly receipt storage.
type Store interface {
	Append(ctx context.Context, rec Record) error
	List(ctx context.Context, q Query) ([]Record, error)
	Close() error
}

// Query filters receipts.
type Query struct {
	Limit      int
	Action     string
	Decision   string
	PathPrefix string
	AgentID    string
}
