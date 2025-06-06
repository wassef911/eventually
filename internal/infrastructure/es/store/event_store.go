package store

import (
	"context"
	"io"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/wassef911/eventually/internal/infrastructure/es"
	"github.com/wassef911/eventually/internal/infrastructure/tracing"
	"github.com/wassef911/eventually/pkg/logger"
)

type eventStore struct {
	log logger.Logger
	db  *esdb.Client
}

func NewEventStore(log logger.Logger, db *esdb.Client) *eventStore {
	return &eventStore{log: log, db: db}
}

func (e *eventStore) SaveEvents(ctx context.Context, streamID string, events []es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "eventStore.SaveEvents")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", streamID))

	eventsData := make([]esdb.EventData, 0, len(events))
	for _, event := range events {
		eventsData = append(eventsData, event.ToEventData())
	}

	_, err := e.db.AppendToStream(ctx, streamID, esdb.AppendToStreamOptions{}, eventsData...)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}

	return nil
}

func (e *eventStore) LoadEvents(ctx context.Context, streamID string) ([]es.Event, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "eventStore.LoadEvents")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", streamID))

	stream, err := e.db.ReadStream(ctx, streamID, esdb.ReadStreamOptions{
		Direction: esdb.Forwards,
		From:      esdb.Revision(1),
	}, 100)
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, err
	}
	defer stream.Close()

	events := make([]es.Event, 0, 100)
	for {
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			tracing.TraceErr(span, err)
			return nil, err
		}
		events = append(events, es.NewEventFromRecorded(event.Event))
	}

	return events, nil
}
