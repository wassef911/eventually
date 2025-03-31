package mongo

import (
	"context"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/wassef911/eventually/internal/api/constants"
	"github.com/wassef911/eventually/internal/delivery/events"
	"github.com/wassef911/eventually/internal/delivery/repository"
	"github.com/wassef911/eventually/internal/infrastructure/es"
	"github.com/wassef911/eventually/internal/infrastructure/tracing"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/logger"
)

type mongoProjection struct {
	log       logger.Logger
	db        *esdb.Client
	config    *config.Config
	mongoRepo repository.MongoRepository
}

func NewOrderProjection(log logger.Logger, db *esdb.Client, mongoRepo repository.MongoRepository, config *config.Config) *mongoProjection {
	return &mongoProjection{log: log, db: db, mongoRepo: mongoRepo, config: config}
}

type Worker func(ctx context.Context, stream *esdb.PersistentSubscription, workerID int) error

func (o *mongoProjection) Subscribe(ctx context.Context, prefixes []string, poolSize int, worker Worker) error {

	err := o.db.CreatePersistentSubscriptionAll(ctx, o.config.Subscriptions.MongoProjectionGroupName, esdb.PersistentAllSubscriptionOptions{
		Filter: &esdb.SubscriptionFilter{Type: esdb.StreamFilterType, Prefixes: prefixes},
	})
	if err != nil {
		if subscriptionError, ok := err.(*esdb.PersistentSubscriptionError); !ok || ok && (subscriptionError.Code != 6) {
			return err
		}
	}

	stream, err := o.db.ConnectToPersistentSubscription(
		ctx,
		constants.EsAll,
		o.config.Subscriptions.MongoProjectionGroupName,
		esdb.ConnectToPersistentSubscriptionOptions{},
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i <= poolSize; i++ {
		g.Go(func() error { return worker(ctx, stream, i) })
	}
	return g.Wait()
}

func (o *mongoProjection) ProcessEvents(ctx context.Context, stream *esdb.PersistentSubscription, workerID int) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		event := stream.Recv()

		switch {
		case event.SubscriptionDropped != nil:
			return errors.Wrap(event.SubscriptionDropped.Error, "subscription dropped")

		case event.EventAppeared != nil:
			return o.processSingleEvent(ctx, stream, event.EventAppeared, workerID)
		}
	}
}

func (o *mongoProjection) processSingleEvent(
	ctx context.Context,
	stream *esdb.PersistentSubscription,
	event *esdb.ResolvedEvent,
	workerID int,
) error {
	o.log.ProjectionEvent(
		constants.MongoProjection,
		o.config.Subscriptions.MongoProjectionGroupName,
		event,
		workerID,
	)

	err := o.When(ctx, es.NewEventFromRecorded(event.Event))
	if err != nil {
		if nackErr := stream.Nack(err.Error(), esdb.Nack_Retry, event); nackErr != nil {
			return errors.Wrap(nackErr, "failed to Nack event")
		}
		return nil
	}

	if ackErr := stream.Ack(event); ackErr != nil {
		return errors.Wrap(ackErr, "failed to Ack event")
	}

	return nil
}

func (o *mongoProjection) When(ctx context.Context, evt es.Event) error {
	ctx, span := tracing.StartProjectionTracerSpan(ctx, "mongoProjection.When", evt)
	defer span.Finish()
	span.LogFields(
		log.String("AggregateID", evt.GetAggregateID()),
		log.String("EventType", evt.GetEventType()),
	)

	handlers := map[string]func(context.Context, es.Event) error{
		events.OrderCreated:           o.onOrderCreate,
		events.OrderPaid:              o.onOrderPaid,
		events.OrderSubmitted:         o.onSubmit,
		events.ShoppingCartUpdated:    o.onShoppingCartUpdate,
		events.OrderCanceled:          o.onCancel,
		events.OrderCompleted:         o.onCompleted,
		events.DeliveryAddressChanged: o.onDeliveryAddressChanged,
	}

	handler, exists := handlers[evt.GetEventType()]
	if !exists {
		o.log.Warnf("(mongoProjection) [When unknown EventType] eventType: {%s}", evt.GetEventType())
		return es.ErrInvalidEventType
	}

	return handler(ctx, evt)
}
