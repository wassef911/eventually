package es

import (
	"context"
	"encoding/json"
)

// SnapshotStore is an interface for an event sourcing snapshot store.
type SnapshotStore interface {
	// SaveSnapshot save aggregate snapshot.
	SaveSnapshot(ctx context.Context, aggregate Aggregate) error

	// GetSnapshot load aggregate snapshot.
	GetSnapshot(ctx context.Context, id string) (*Snapshot, error)
}

// Snapshot Event Sourcing Snapshotting is an optimisation that reduces time spent on reading event from an event store.
type Snapshot struct {
	ID      string        `json:"id"`
	Type    AggregateType `json:"type"`
	State   []byte        `json:"state"`
	Version uint64        `json:"version"`
}

// NewSnapshotFromAggregate create new snapshot from the Aggregate state.
func NewSnapshotFromAggregate(aggregate Aggregate) (*Snapshot, error) {

	aggregateBytes, err := json.Marshal(aggregate)
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		ID:      aggregate.GetID(),
		Type:    aggregate.GetType(),
		State:   aggregateBytes,
		Version: uint64(aggregate.GetVersion()),
	}, nil
}
