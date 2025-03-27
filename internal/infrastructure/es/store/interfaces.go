package store

import (
	"context"

	"github.com/wassef911/eventually/internal/infrastructure/es"
)

// AggregateStore is responsible for loading and saving aggregates.
type AggregateStore interface {
	// Load loads the most recent version of an aggregate to provided  into params aggregate with a type and id.
	Load(ctx context.Context, aggregate es.Aggregate) error

	// Save saves the uncommitted events for an aggregate.
	Save(ctx context.Context, aggregate es.Aggregate) error

	// Exists check aggregate exists by id.
	Exists(ctx context.Context, streamID string) error

	//EventStore
	//SnapshotStore
}

// EventStore is an interface for an event sourcing event store.
type EventStore interface {
	// SaveEvents appends all events in the event stream to the store.
	SaveEvents(ctx context.Context, streamID string, events []es.Event) error

	// LoadEvents loads all events for the aggregate id from the store.
	LoadEvents(ctx context.Context, streamID string) ([]es.Event, error)
}
