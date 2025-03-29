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
	cfg       *config.Config
	mongoRepo repository.OrderMongoRepository
}

func NewOrderProjection(log logger.Logger, db *esdb.Client, mongoRepo repository.OrderMongoRepository, cfg *config.Config) *mongoProjection {
	return &mongoProjection{log: log, db: db, mongoRepo: mongoRepo, cfg: cfg}
}

type Worker func(ctx context.Context, stream *esdb.PersistentSubscription, workerID int) error

func (o *mongoProjection) Subscribe(ctx context.Context, prefixes []string, poolSize int, worker Worker) error {

	err := o.db.CreatePersistentSubscriptionAll(ctx, o.cfg.Subscriptions.MongoProjectionGroupName, esdb.PersistentAllSubscriptionOptions{
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
		o.cfg.Subscriptions.MongoProjectionGroupName,
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
		event := stream.Recv()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if event.SubscriptionDropped != nil {
			return errors.Wrap(event.SubscriptionDropped.Error, "Subscription Dropped")
		}

		if event.EventAppeared != nil {
			o.log.ProjectionEvent(constants.MongoProjection, o.cfg.Subscriptions.MongoProjectionGroupName, event.EventAppeared, workerID)

			err := o.When(ctx, es.NewEventFromRecorded(event.EventAppeared.Event))
			if err != nil {
				if err := stream.Nack(err.Error(), esdb.Nack_Retry, event.EventAppeared); err != nil {
					return errors.Wrap(err, "stream.Nack")
				}
			}

			err = stream.Ack(event.EventAppeared)
			if err != nil {
				return errors.Wrap(err, "stream.Ack")
			}
		}
	}
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
